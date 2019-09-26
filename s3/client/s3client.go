package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	minio "github.com/minio/minio-go"
)

type S3Client struct {
	Endpoint         string
	AccessKeyID      string
	SecretAccessKey  string
	BucketName       string
	Location         string
	SignatureVersion string
	UploadName       string
	ObjectName       string
	DownloadName     string
	UseSSL           bool
	S3               *minio.Client

	FoundBucket bool
}

func NewS3Client(endpoint, accessKeyID, secretAccessKey, bucketName, location,
	signatureVersion, uploadName, objectName, downloadName string, useSSL bool) *S3Client {
	return &S3Client{
		Endpoint:         endpoint,
		AccessKeyID:      accessKeyID,
		SecretAccessKey:  secretAccessKey,
		BucketName:       bucketName,
		Location:         location,
		SignatureVersion: signatureVersion,
		UploadName:       uploadName,
		ObjectName:       objectName,
		DownloadName:     downloadName,
		UseSSL:           useSSL,
	}
}

func (c *S3Client) PrintParam() {
	fmt.Println("\nS3Client: ")
	fmt.Printf(" -e\tEndpoint: '%s'\n", c.Endpoint)
	fmt.Printf(" -ak\tAccessKey: '%s'\n", c.AccessKeyID)
	fmt.Printf(" -sk\tSecretKey: '%s'\n", c.SecretAccessKey)
	fmt.Printf(" -b\tBucketName: '%s'\n", c.BucketName)
	fmt.Println(" -ssl\tUseSSL: ", c.UseSSL)
	fmt.Printf(" -l\tLocation: '%s'\n", c.Location)
	fmt.Printf(" -v\tSignatureVersion: '%s'\n", c.SignatureVersion)
	fmt.Printf(" -u\tUploadFileName: '%s'\n", c.UploadName)
	fmt.Printf(" -o\tObjectName: '%s'\n", c.ObjectName)
	fmt.Printf(" -d\tDownloadFileName: '%s'\n", c.DownloadName)
}

func (c *S3Client) Connect() (err error) {

	if c.SignatureVersion == "v2" {
		c.S3, err = minio.NewV2(c.Endpoint, c.AccessKeyID, c.SecretAccessKey, c.UseSSL)
		if err != nil {
			return
		}
	} else {
		c.S3, err = minio.New(c.Endpoint, c.AccessKeyID, c.SecretAccessKey, c.UseSSL)
		if err != nil {
			return
		}
	}
	return
}

func (c *S3Client) CheckBucket() (err error) {

	/* BucketExists */
	fmt.Println("\nBucketExists: ", c.BucketName)
	c.FoundBucket, err = c.S3.BucketExists(c.BucketName)
	if err != nil {
		fmt.Println(err)
	}

	/* MakeBucket */
	if !c.FoundBucket {
		fmt.Println("\nMakeBucket: ", c.BucketName)
		err = c.S3.MakeBucket(c.BucketName, c.Location)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	return
}

func (c *S3Client) ListBucket() (err error) {

	/* ListBuckets */
	fmt.Println("\nListBucket: ")
	buckets, err := c.S3.ListBuckets()
	if err != nil {
		fmt.Println(err)
	}
	for _, bucket := range buckets {
		fmt.Println(bucket.Name)
	}
	return
}

func (c *S3Client) PutObject() (err error) {
	fmt.Printf("\nPutObject: %s -> %s/%s\n", c.UploadName, c.BucketName, c.ObjectName)
	file, err := os.Open(c.UploadName)
	if err != nil {
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return
	}

	n, err := c.S3.PutObject(c.BucketName, c.ObjectName, file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return
	}
	fmt.Println("Successfully uploaded bytes:", n)
	return
}

func (c *S3Client) HeadObject() (err error) {

	/* HeadObject */
	fmt.Printf("\nHeadObject: %s/%s\n", c.BucketName, c.ObjectName)
	objInfo, err := c.S3.StatObject(c.BucketName, c.ObjectName, minio.StatObjectOptions{})
	if err != nil {
		return
	}
	fmt.Println(objInfo)
	return
}

func (c *S3Client) ListObject() (err error) {

	/* ListObjects */
	fmt.Printf("\nListObjects: %s/\n", c.BucketName)
	doneCh := make(chan struct{}) // Create a done channel to control 'ListObjects' go routine.
	defer close(doneCh)           // Indicate to our routine to exit cleanly upon return.
	isRecursive := true
	objectCh := c.S3.ListObjects(c.BucketName, "", isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			err = object.Err
		} else {
			fmt.Println(object)
		}
	}
	return
}

func (c *S3Client) GetObject() (err error) {
	fmt.Printf("\nGetObject: %s/%s -> %s\n", c.BucketName, c.ObjectName, c.DownloadName)
	object, err := c.S3.GetObject(c.BucketName, c.ObjectName, minio.GetObjectOptions{})
	if err != nil {
		return
	}
	defer object.Close()

	localFile, err := os.Create(c.DownloadName)
	if err != nil {
		return
	}
	defer localFile.Close()

	var n int64
	if n, err = io.Copy(localFile, object); err != nil {
		return
	}
	fmt.Println("Successfully download bytes:", n)
	return
}

func (c *S3Client) RemoveObject() (err error) {
	fmt.Printf("\nRemoveObject: %s/%s\n", c.BucketName, c.ObjectName)
	err = c.S3.RemoveObject(c.BucketName, c.ObjectName)
	if err != nil {
		return
	}
	return
}

func (c *S3Client) RemoveBucket() (err error) {
	if !c.FoundBucket {
		fmt.Printf("\nRemoveObject: %s\n", c.BucketName)
		err = c.S3.RemoveBucket(c.BucketName)
		if err != nil {
			return
		}
	}
	return
}

func main() {
	var endpoint, accessKeyID, secretAccessKey, bucketName, location, signatureVersion, uploadName, objectName, downloadName string
	var useSSL bool

	/* paramParse */
	flag.Bool("?|h", false, "./s3client -e 123.177.21.80:8004 -ak Q3AM3UQ867SPQQA43P2F -sk zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG -b bucket1 -u ./up.zip -o dir1/up.zip -d ./down.zip")
	flag.StringVar(&endpoint, "e", "", "Endpoint")
	flag.StringVar(&accessKeyID, "ak", "", "AccessKey")
	flag.StringVar(&secretAccessKey, "sk", "", "SecretKey")
	flag.StringVar(&bucketName, "b", "bucket01", "BucketName")
	flag.BoolVar(&useSSL, "ssl", false, "UseSSL")
	flag.StringVar(&location, "l", "us-east-1", "Location")
	flag.StringVar(&signatureVersion, "v", "v4", "SignatureVersion")
	flag.StringVar(&uploadName, "u", "./up.zip", "UploadFileName")
	flag.StringVar(&objectName, "o", "up.zip", "ObjectName")
	flag.StringVar(&downloadName, "d", "./down.zip", "DownloadFileName")

	flag.Parse()
	parsed := flag.Parsed()
	if parsed {
		if endpoint == "" {
			flag.Usage()
			return
		}
	}

	/* NewS3Clent */
	client := NewS3Client(endpoint, accessKeyID, secretAccessKey, bucketName, location,
		signatureVersion, uploadName, objectName, downloadName, useSSL)
	client.PrintParam()

	/* s3Client */
	var err error
	err = client.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* CheckBucket */
	err = client.CheckBucket()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* PutObject */
	err = client.PutObject()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* HeadObject */
	err = client.HeadObject()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* ListObject */
	err = client.ListObject()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* GetObject */
	err = client.GetObject()
	if err != nil {
		fmt.Println(err)
		return
	}

	/* RemoveObject*/
	err = client.RemoveObject()
	if err != nil {
		fmt.Println(err)
	}

	/* RemoveBucket */
	err = client.RemoveBucket()
	if err != nil {
		fmt.Println(err)
	}

	return
}
