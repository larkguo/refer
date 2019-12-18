
/*======================copy.go======================*/
package copy
import (
	"context"
	"github.com/rclone/rclone/cmd"
)
func init() {
	cmd.Root.AddCommand(commandDefinition)
	cmdFlags := commandDefinition.Flags()
	flags.BoolVarP(cmdFlags, &createEmptySrcDirs, "create-empty-src-dirs", "", createEmptySrcDirs, "Create empty source dirs on destination after copy")
}
var commandDefinition = &cobra.Command{
	Use:   "copy source:path dest:path",
	Short: `Copy files from source to dest, skipping already copied`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(2, 2, command, args)
		fsrc, srcFileName, fdst := cmd.NewFsSrcFileDst(args)
		cmd.Run(true, true, command, func() error {
			if srcFileName == "" {
				return sync.CopyDir(context.Background(), fdst, fsrc, createEmptySrcDirs)
			}
			return operations.CopyFile(context.Background(), fdst, fsrc, srcFileName, srcFileName)
		})
	},
}
/*======================operations.go======================*/
package operations
import ( "github.com/rclone/rclone/fs")
// Copy src object to dst or f if nil. If dst is nil then it uses remote as the name of the new object.
// It returns the destination object if possible.  Note that this may be nil.
func Copy(ctx context.Context, f fs.Fs, dst fs.Object, remote string, src fs.Object) (newDst fs.Object, err error) {
	tr := accounting.Stats(ctx).NewTransfer(src)
	defer func() {
		tr.Done(err)
	}()
	newDst = dst
	if fs.Config.DryRun {
		fs.Logf(src, "Not copying as --dry-run")
		return newDst, nil
	}
	maxTries := fs.Config.LowLevelRetries
	tries := 0
	doUpdate := dst != nil
	// work out which hash to use - limit to 1 hash in common
	var common hash.Set
	hashType := hash.None
	if !fs.Config.IgnoreChecksum {
		common = src.Fs().Hashes().Overlap(f.Hashes())
		if common.Count() > 0 {
			hashType = common.GetOne()
			common = hash.Set(hashType)
		}
	}
	hashOption := &fs.HashesOption{Hashes: common}
	var actionTaken string
	for {
		var in0 io.ReadCloser
		in0, _ = newReOpen(ctx, src, hashOption, nil, fs.Config.LowLevelRetries)
		in := tr.Account(in0).WithBuffer() // account and buffer the transfer 分配buffer
		var wrappedSrc fs.ObjectInfo = src
		// We try to pass the original object if possible
		if src.Remote() != remote {
			wrappedSrc = &overrideRemoteObject{Object: src, remote: remote}
		}
		if doUpdate {
			actionTaken = "Copied (replaced existing)" 
			err = dst.Update(ctx, in, wrappedSrc, hashOption) 
		} else {
			actionTaken = "Copied (new)" 
			dst, err = f.Put(ctx, in, wrappedSrc, hashOption) // 上传,回调interface接口Put具体实现(如S3 Put)
		}
		closeErr := in.Close()
		if err == nil {
			newDst = dst
			err = closeErr
		}

		tries++
		if tries >= maxTries {
			break
		}
		// Retry if err returned a retry error
		if fserrors.IsRetryError(err) || fserrors.ShouldRetry(err) {
			fs.Debugf(src, "Received error: %v - low level retry %d/%d", err, tries, maxTries)
			continue
		}
		// otherwise finish
		break
	}
	return newDst, err
}
/*======================s3.go======================*/
package s3
import (  "github.com/aws/aws-sdk-go/aws")
// Put the Object into the bucket
func (f *Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	// Temporary Object under construction
	fs := &Object{
		fs:     f,
		remote: src.Remote(),
	}
	return fs, fs.Update(ctx, in, src, options...)
}
// Update the Object from in with modTime and size
func (o *Object) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	bucket, bucketPath := o.split()
	err := o.fs.makeBucket(ctx, bucket)
	if err != nil {
		return err
	}
	modTime := src.ModTime(ctx)
	size := src.Size()
	var uploader *s3manager.Uploader

	// Set the mtime in the meta data
	metadata := map[string]*string{
		metaMtime: aws.String(swift.TimeToFloatString(modTime)),
	}

	mimeType := fs.MimeType(ctx, src)// Guess the content type
	req := s3.PutObjectInput{
		Bucket:      &bucket,
		ACL:         &o.fs.opt.ACL,
		Key:         &bucketPath,
		ContentType: &mimeType,
		Metadata:    metadata,
	}

	putObj, _ := o.fs.c.PutObjectRequest(&req)// Create the request

	// Sign it so we can upload using a presigned request.
	// Note the SDK doesn't currently support streaming to
	// PutObject so we'll use this work-around.
	url, headers, err := putObj.PresignRequest(15 * time.Minute)
	if err != nil {
		return errors.Wrap(err, "s3 upload: sign request")
	}

	// create the vanilla http request
	httpReq, err := http.NewRequest("PUT", url, in)
	if err != nil {
		return errors.Wrap(err, "s3 upload: new request")
	}
	httpReq = httpReq.WithContext(ctx) // go1.13 can use NewRequestWithContext

	// set the headers we signed and the length
	httpReq.Header = headers
	httpReq.ContentLength = size

	err = o.fs.pacer.CallNoRetry(func() (bool, error) {
		resp, err := o.fs.srv.Do(httpReq)
		if err != nil {
			return o.fs.shouldRetry(err)
		}
		body, err := rest.ReadBody(resp)
		if err != nil {
			return o.fs.shouldRetry(err)
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 299 {
			return false, nil
		}
		err = errors.Errorf("s3 upload: %s: %s", resp.Status, body)
		return fserrors.ShouldRetryHTTP(resp, retryErrorCodes), err
	})

	// Read the metadata from the newly created object
	o.meta = nil // wipe old metadata
	err = o.readMetaData(ctx)
	return err
}
/*======================client.go======================*/
package http
func (c *Client) Do(req *Request) (retres *Response, reterr error) {
	var (
		deadline      = c.deadline()
		resp          *Response
	)
	var err error
	var didTimeout func() bool

	//发送请求到服务端，并获取响应信息resp
	if resp, didTimeout, err = c.send(req, deadline); err != nil {
		// c.send() always closes req.Body
		reqBodyClosed = true
		if !deadline.IsZero() && didTimeout() { //已超时
			err = &httpError{
				// TODO: early in cycle: s/Client.Timeout exceeded/timeout or context cancelation/
				err:     err.Error() + " (Client.Timeout exceeded while awaiting headers)",
				timeout: true,
			}
		}
		return nil, uerr(err)
	}
	req.closeBody()
}
/*======================transport.go======================*/
package http
/*
 roundtrip处理读写的大致流程涉及三个goroutine，其实逻辑很简单清晰：
 dialconn新建TCP连接时，然后开始readLoop和writeLoop
 通过getConn获得连接后，roundtrip将req和writeErrCh发给writeLoop，writeLoop把发请求的结果通过writeErrCh通知roundtrip这个主协程
 同时roundtrip将req和responseAndErrorCh发给readLoop，readLoop把相应和error通知主协程。
 */
 type persistConn struct {
	alt RoundTripper
	t         *Transport
	cacheKey  connectMethodKey
	conn      net.Conn
	tlsState  *tls.ConnectionState
	br        *bufio.Reader       // from conn
	bw        *bufio.Writer       // to conn
	nwrite    int64               // bytes written
	//roundTrip往这个chan里写入request,readLoop从这个chan读取request
	reqch     chan requestAndChan // written by roundTrip; read by readLoop 
	//roundTrip往这个chan里写入request和writeErrCh,writeLoop从这个chan读取request写入conn连接里,并写入err到writeErrCh
	writech   chan writeRequest   // written by roundTrip; read by writeLoop 
	closech   chan struct{}       // closed when conn closed
	sawEOF    bool  // whether we've seen EOF from conn; owned by readLoop 判断body是否读取完
	writeErrCh chan error // writeLoop 写入err的chan
	writeLoopDone chan struct{} // closed when write loop ends. writeLoop 结束的时候关闭
 }
 func (t *Transport) dialConn(cm connectMethod) (*persistConn, error) {
    pconn := &persistConn{
        t:          t,
        cacheKey:   cm.key(),
        reqch:      make(chan requestAndChan, 1),
        writech:    make(chan writeRequest, 1),
        closech:    make(chan struct{}),
        writeErrCh: make(chan error, 1),
    }
    conn, err := t.dial("tcp", cm.addr())
    pconn.conn = conn

    pconn.br = bufio.NewReader(noteEOFReader{pconn.conn, &pconn.sawEOF})
    pconn.bw = bufio.NewWriter(pconn.conn)
    go pconn.readLoop()
    go pconn.writeLoop()
    return pconn, nil
}

 //roundTrip其实就是把req写到writeCh,即writeLoop开始往conn上发request,
 //同时把resc这个用来收集response和error的ch通过pc.reqch上发给连接的readLoop.package
 //然后开始等结果,若写req错误,则返回,若读循环resc有结果也返回
 func (pc *persistConn) roundTrip(req *transportRequest) (resp *Response, err error) {
    writeErrCh := make(chan error, 1)
    pc.writech <- writeRequest{req, writeErrCh} //把req写到writeCh,writeLoop开始向conn发request

    resc := make(chan responseAndError, 1)
    pc.reqch <- requestAndChan{req.Request, resc, requestedGzip} //把resc这个channel用来收集response
    var re responseAndError
WaitResponse:
    for {
        select {
        case err := <-writeErrCh:
            if err != nil {
                re = responseAndError{nil, err}
                pc.close()
                break WaitResponse
            }
        case re = <-resc: //接收到readLoop返回的response
            break WaitResponse
        }
    }
    return re.res, re.err
}

//writeLoop阻塞在<-pc.writeCh,直到roundTrip开始传入req,于是往pconn上写请求.
//处理完一次req并返回结果后,writeLoop重新阻塞在pc.writeCh直到这个连接被复用,有另一个http请求发送.
func (pc *persistConn) writeLoop() {
    for {
        select {
        case wr := <-pc.writech: 
            err := wr.req.Request.write(pc.bw, pc.isProxy, wr.req.extra) // 写入TCP流
            if err == nil {
                err = pc.bw.Flush()
            }
            pc.writeErrCh <- err // to the body reader, which might recycle us
            wr.ch <- err         // to the roundTrip function
        case <-pc.closech:
            return
        }
    }
}
//readLoop阻塞在pc.reqch,直到roundTrip开始,readLoop开始读取pconn的response,并把结果返回给主循环.
func (pc *persistConn) readLoop() {
    alive := true
    for alive {
        pb, err := pc.br.Peek(1)
        rc := <-pc.reqch
        ...
        resp, err = ReadResponse(pc.br, rc.req) 
        rc.ch <- responseAndError{resp, err}
    }
    pc.close()
}
