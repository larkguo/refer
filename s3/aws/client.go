package main

import (
	"archive/zip"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/minio/minio-go/v7"
	credentials2 "github.com/minio/minio-go/v7/pkg/credentials"
)

type FakeWriterAt struct {
	w io.Writer
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	return fw.w.Write(p) // ignore 'offset' because we forced sequential downloads
}
func PutZipFile(ctx aws.Context, sess *session.Session, file *s3.GetObjectInput, result *s3manager.UploadInput) error {
	downloader := s3manager.NewDownloader(sess)
	uploader := s3manager.NewUploader(sess)
	pr, pw := io.Pipe()

	// Create zip.Write which will writes to pipes
	zipWriter := zip.NewWriter(pw)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() { // Run downloader
		// We need to close our zip.Writer and also pipe writer
		// zip.Writer doesn't close underylying writer
		defer func() {
			wg.Done()
			zipWriter.Close()
			pw.Close()
		}()
		// Sequantially downloads each file to writer from zip.Writer
		w, err := zipWriter.Create(path.Base(*file.Key))
		if err != nil {
			fmt.Println(err)
		}
		_, err = downloader.DownloadWithContext(ctx, FakeWriterAt{w}, file,
			func(d *s3manager.Downloader) {
				d.Concurrency = 1
			})
		if err != nil {
			fmt.Println(err)
		}
	}()
	go func() { // Run uploader
		defer wg.Done()
		// Upload the file, body is `io.Reader` from pipe
		result.Body = pr
		_, err := uploader.UploadWithContext(ctx, result)
		if err != nil {
			fmt.Println(err)
		}
	}()
	wg.Wait()
	return nil
}
func PutMd5File(url, accesskey, secretkey, bucket, filename, region *string, pathstyle *bool) error {
	var useSSL bool
	var endpoint string
	var lookup minio.BucketLookupType
	if *pathstyle {
		lookup = minio.BucketLookupPath
	} else {
		lookup = minio.BucketLookupDNS
	}
	ep := *url
	if strings.HasPrefix(ep, "https://") {
		useSSL = true
		endpoint = ep[8:]
	} else if strings.HasPrefix(ep, "http://") {
		useSSL = false
		endpoint = ep[7:]
	}

	// Initialize minio client object.
	s3Client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials2.NewStaticV4(*accesskey, *secretkey, ""),
		Secure:       useSSL,
		Region:       *region,
		BucketLookup: lookup,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	file, err := os.Open(*filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return err
	}

	uploadInfo, err := s3Client.PutObject(context.Background(),
		*bucket, *filename, file, fileStat.Size(),
		minio.PutObjectOptions{
			ContentType:    "application/octet-stream",
			SendContentMd5: true,
			PartSize:       5 * 1024 * 1024, // 5MB per part
		})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)
	return nil
}

func PutFile(ctx aws.Context, sess *session.Session, bucket *string, filename *string) error {
	file, err := os.Open(*filename)
	if err != nil {
		fmt.Println("Unable to open file " + *filename)
		return err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: bucket,
		Key:    filename,
		Body:   file,
	})
	return err
}
func DownloadObject(ctx aws.Context, sess *session.Session, bucket *string, filename *string) error {
	path := filepath.Dir(*filename)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(*filename)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.DownloadWithContext(ctx, file,
		&s3.GetObjectInput{
			Bucket: bucket,
			Key:    filename,
		})
	return err
}
func DeleteItem(ctx aws.Context, sess *session.Session, bucket *string, item *string) error {
	svc := s3.New(sess)
	_, err := svc.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: bucket,
		Key:    item,
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilObjectNotExistsWithContext(ctx, &s3.HeadObjectInput{
		Bucket: bucket,
		Key:    item,
	})
	return err
}
func ListObjectsPages(ctx aws.Context, sess *session.Session, bucket, prefix, delimiter *string, maxkeys *int64) error {
	var total int64

	svc := s3.New(sess)
	err := svc.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
		Bucket:    aws.String(*bucket),
		MaxKeys:   aws.Int64(*maxkeys),
		Prefix:    aws.String(*prefix),
		Delimiter: aws.String(*delimiter),
	}, func(p *s3.ListObjectsOutput, lastPage bool) bool {
		for _, object := range p.Contents {
			fmt.Println("Name:          ", *object.Key)
			fmt.Println("Last modified: ", *object.LastModified)
			fmt.Println("ETag:          ", *object.ETag)
			fmt.Println("Size:          ", *object.Size)
			fmt.Println("Storage class: ", *object.StorageClass)
			fmt.Println("")
			total++
		}
		fmt.Println("LastPage:", lastPage, " Total:", total, "\n")

		return true // continue paging
	})
	return err
}
func GetAllBuckets(ctx aws.Context, sess *session.Session) (*s3.ListBucketsOutput, error) {
	svc := s3.New(sess)
	result, err := svc.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	return result, err
}
func RestoreItem(ctx aws.Context, sess *session.Session, bucket *string, item *string, days int64) error {
	svc := s3.New(sess)
	_, err := svc.RestoreObjectWithContext(ctx, &s3.RestoreObjectInput{
		Bucket: bucket,
		Key:    item,
		RestoreRequest: &s3.RestoreRequest{
			Days: aws.Int64(days),
		},
	})
	return err
}
func ConfirmBucketItemExists(ctx aws.Context, sess *session.Session, bucket *string, item *string) error {
	svc := s3.New(sess)
	object, err := svc.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: bucket,
		Key:    item,
	})
	if err != nil {
		return err
	}
	fmt.Println(object)
	return nil
}
func ParseHeadObject(ctx aws.Context, svc *s3.S3, bucket *string, filename *string) (standard int, ongoing int, err error) {
	standard = -1
	ongoing = -1

	object, err := svc.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: bucket,
		Key:    filename,
	})
	if err != nil {
		return standard, ongoing, err
	}

	if object.StorageClass != nil {
		if strings.Contains(*object.StorageClass, "GLACIER") {
			standard = 0
			if object.Restore != nil && strings.Contains(*object.Restore, "ongoing-request=\"false\"") {
				ongoing = 0
			} else if object.Restore != nil && strings.Contains(*object.Restore, "ongoing-request=\"true\"") {
				ongoing = 1
			} else {
				ongoing = -1
			}
		} else {
			standard = 1
		}
	}
	fmt.Println("Standard:", standard, " ongoing-request:", ongoing)
	return standard, ongoing, err
}
func GetObject(ctx aws.Context, sess *session.Session, bucket *string, filename *string) error {
	svc := s3.New(sess)

ParseHeadObject:
	standard, glacier_ongoing, err := ParseHeadObject(ctx, svc, bucket, filename)
	if err != nil {
		return err
	}
	if standard == 0 {
		if glacier_ongoing == -1 {
			_, err = svc.RestoreObjectWithContext(ctx, &s3.RestoreObjectInput{
				Bucket: bucket,
				Key:    filename,
				RestoreRequest: &s3.RestoreRequest{
					Days: aws.Int64(1),
				},
			})
			if err != nil {
				return err
			}
			select {
			case <-time.After(60 * time.Second):
				goto ParseHeadObject
			}
		}
		if glacier_ongoing == 1 {
			select {
			case <-time.After(5 * time.Second):
				goto ParseHeadObject
			}
		}
	}

	err = os.MkdirAll(filepath.Dir(*filename), os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(*filename)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.DownloadWithContext(ctx, file,
		&s3.GetObjectInput{
			Bucket: bucket,
			Key:    filename,
		})
	return err
}

func GetObject2(ctx aws.Context, sess *session.Session, bucket *string, filename *string) error {
	svc := s3.New(sess)
	var err error
	ok := make(chan int)

	go func() {
	ParseHeadObject:
		standard, glacier_ongoing, err := ParseHeadObject(ctx, svc, bucket, filename)
		if err != nil {
			ok <- 0
			return
		}
		if standard == 1 {
			ok <- 1
			return
		}
		if standard == 0 {
			if glacier_ongoing == -1 {
				_, err = svc.RestoreObjectWithContext(ctx, &s3.RestoreObjectInput{
					Bucket: bucket,
					Key:    filename,
					RestoreRequest: &s3.RestoreRequest{
						Days: aws.Int64(1),
					},
				})
				if err != nil {
					ok <- 0
					return
				}
				select {
				case <-time.After(60 * time.Second):
					goto ParseHeadObject
				}
			}
			if glacier_ongoing == 1 {
				select {
				case <-time.After(5 * time.Second):
					goto ParseHeadObject
				}
			}
			if glacier_ongoing == 0 {
				ok <- 1
				return
			}
		}
	}()

	restore := -1
	select {
	case restore := <-ok:
		fmt.Println("Restore ok:", restore)
		close(ok)
	}
	if restore == 1 {
		err = os.MkdirAll(filepath.Dir(*filename), os.ModePerm)
		if err != nil {
			return err
		}
		file, err := os.Create(*filename)
		if err != nil {
			return err
		}
		defer file.Close()

		downloader := s3manager.NewDownloader(sess)
		_, err = downloader.DownloadWithContext(ctx, file,
			&s3.GetObjectInput{
				Bucket: bucket,
				Key:    filename,
			})
	}
	return err
}

func main() {
	var err error
	handle := flag.String("h", "list", "The handle of up|down|get|get2|del|head|list|res|md5|zip")
	bucket := flag.String("b", "ehlbucket01", "Bucket ")
	filename := flag.String("f", "", "The name of the file")
	accesskey := flag.String("a", "LTAI4GFBJd1WwEMAjx3C1ZtS0", "AccessKey")
	secretkey := flag.String("s", "97KS9V5PjX7pQJJqQLiFK3sRS8MW060", "SecretKey")
	endpoint := flag.String("e", "http://oss-cn-beijing.aliyuncs.com:80", "Endpoint( http(s):// )")
	region := flag.String("r", "us-east-1", "Region ")
	pathstyle := flag.Bool("v", false, "The Path-Style(1) or Virtual-Hosted-Style(0) ")
	prefix := flag.String("p", "", "ListObjects Prefix ")
	maxkeys := flag.Int64("m", 1000, "ListObjects MaxKeys ")
	delimiter := flag.String("d", "", "ListObjects Delimiter ")
	flag.Parse()
	if *accesskey == "" || *secretkey == "" || *endpoint == "" {
		fmt.Println("Specify a Handle(-h UP|DOWN|GET|DEL|HEAD|LIST|RES|MD5|ZIP) Bucket (-b BUCKET) filename (-f FILENAME) AccessKey(-a AK) SecretKey(-s SK) Endpoint(-e ENDPOINT) PathStyle(-v 1|0) Region(-r REGION) Prefix(-p PREFIX) Maxkeys(-m MAXKEYS) Delimiter(-d DELIMITER) ")
		return
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(*accesskey, *secretkey, ""),
		Endpoint:         aws.String(*endpoint),
		Region:           aws.String(*region),
		HTTPClient:       &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		S3ForcePathStyle: aws.Bool(*pathstyle),
	}))

	ctx := context.Background()
	start := time.Now()

	if *handle == "list" || *handle == "LIST" {
		if *bucket == "" { // buckets
			result, err := GetAllBuckets(ctx, sess)
			if err != nil {
				fmt.Println("Got error retrieving buckets:")
				fmt.Println(err)
				return
			}
			fmt.Println("Buckets:")
			for _, bucket := range result.Buckets {
				fmt.Println(*bucket.Name + ": " + bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
			}

		} else { // objects
			err := ListObjectsPages(ctx, sess, bucket, prefix, delimiter, maxkeys)
			if err != nil {
				fmt.Println("Got error retrieving list of objects:")
				fmt.Println(err)
				return
			}
		}
	} else if *handle == "get" || *handle == "GET" {
		err = GetObject(ctx, sess, bucket, filename)
		if err != nil {
			fmt.Println("Got error get " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Got " + *filename)
	} else if *handle == "get2" || *handle == "GET2" {

		err = GetObject2(ctx, sess, bucket, filename)
		if err != nil {
			fmt.Println("Got error get2 " + *filename + ":")
			fmt.Println(err)
			return
		}
	} else if *handle == "up" || *handle == "UP" {
		err = PutFile(ctx, sess, bucket, filename)
		if err != nil {
			fmt.Println("Got error upload " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Uploaded " + *filename)
	} else if *handle == "md5" || *handle == "MD5" {
		err = PutMd5File(endpoint, accesskey, secretkey, bucket, filename, region, pathstyle)
		if err != nil {
			fmt.Println("Got error md5 upload " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Uploaded " + *filename)
	} else if *handle == "zip" || *handle == "ZIP" {
		PutZipFile(ctx, sess, &s3.GetObjectInput{
			Bucket: bucket,
			Key:    filename,
		}, &s3manager.UploadInput{
			Bucket: bucket,
			Key:    aws.String(*filename + ".zip"),
		})
	} else if *handle == "down" || *handle == "DOWN" {
		err = DownloadObject(ctx, sess, bucket, filename)
		if err != nil {
			fmt.Println("Got error download " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Downloaded " + *filename)
	} else if *handle == "del" || *handle == "DEL" {
		err = DeleteItem(ctx, sess, bucket, filename)
		if err != nil {
			fmt.Println("Got error delete " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Deleted " + *filename)
	} else if *handle == "head" || *handle == "HEAD" {
		err = ConfirmBucketItemExists(ctx, sess, bucket, filename)
		if err != nil {
			fmt.Println("Got error head " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Headed " + *filename)
	} else if *handle == "res" || *handle == "RES" {

		err := RestoreItem(ctx, sess, bucket, filename, 7)
		if err != nil {
			fmt.Println("Got error restore " + *filename + ":")
			fmt.Println(err)
			return
		}
		fmt.Println("Restored " + *filename + " to " + *bucket)
	}

	elapsed := time.Since(start)
	fmt.Println("This function took:", elapsed)
}
