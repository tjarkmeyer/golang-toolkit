package gcp

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/api/iterator"
)

type Config struct {
	PROJECT_ID                     string `default:"" envconfig:"PROJECT_ID"`
	BUCKET_NAME                    string `default:"" envconfig:"BUCKET_NAME"`
	GOOGLE_APPLICATION_CREDENTIALS string `default:"/etc/gcp/sa_credentials.json" envconfig:"GOOGLE_APPLICATION_CREDENTIALS"`
}

type Client struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

// New - new client to upload/download files
func New(bucketName, projectID, uploadPath string) (*Client, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return &Client{}, err
	}

	return &Client{
		cl:         client,
		bucketName: bucketName,
		projectID:  projectID,
		uploadPath: uploadPath,
	}, nil
}

// Upload - uploads data, filename should contain filetype
func Upload(ctx context.Context, filename string, data []byte, client *Client) error {
	return client.uploadFile(ctx, data, filename)
}

// Download - downloads data, filename should contain filetype
func Download(ctx context.Context, filename string, client *Client) ([]byte, error) {
	return client.downloadFile(ctx, filename)
}

// Delete - deletes data, filename should contain filetype
func Delete(ctx context.Context, filename string, client *Client) error {
	return client.deleteFile(ctx, filename)
}

// DeleteAllByPrefix - deletes all objects that matches the prefix name
func DeleteAllByPrefix(ctx context.Context, prefix string, client *Client) error {
	return client.deleteFilesByPrefix(ctx, prefix)
}

// ListAll - retruns an iterator to list all objects that match the prefix
func ListAllFilenames(ctx context.Context, prefix string, client *Client) ([]string, error) {
	return client.listFilenames(ctx, prefix)
}

// DownloadFirstByPrefix - retruns an iterator to list all objects that match the prefix
func DownloadFirstByPrefix(ctx context.Context, prefix string, client *Client) ([]byte, string, error) {
	return client.donwloadFirstFileByPrefix(ctx, prefix)
}

// uploadFile uploads an object
func (c *Client) uploadFile(ctx context.Context, file []byte, object string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer
	if !strings.HasPrefix(object, c.uploadPath) {
		object = c.uploadPath + object
	}
	wc := c.cl.Bucket(c.bucketName).Object(object).NewWriter(ctx)
	if _, err := wc.Write(file); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

// downloadFile downloads an object
func (c *Client) downloadFile(ctx context.Context, object string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Donwload an object with storage.Reader
	if !strings.HasPrefix(object, c.uploadPath) {
		object = c.uploadPath + object
	}
	rc, err := c.cl.Bucket(c.bucketName).Object(object).NewReader(ctx)
	if err != nil {
		return []byte{}, err
	}
	slurp, err := io.ReadAll(rc)
	rc.Close()

	return slurp, err
}

// deleteFile downloads an object
func (c *Client) deleteFile(ctx context.Context, object string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Get an object and delete it
	if !strings.HasPrefix(object, c.uploadPath) {
		object = c.uploadPath + object
	}
	o := c.cl.Bucket(c.bucketName).Object(object)
	err := o.Delete(ctx)

	return err
}

// deleteFilesByPrefix - deletes all files that match the prefix
func (c *Client) deleteFilesByPrefix(ctx context.Context, prefix string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Get the iterator for matching objects and delete each of them
	if !strings.HasPrefix(prefix, c.uploadPath) {
		prefix = c.uploadPath + prefix
	}
	it := c.cl.Bucket(c.bucketName).Objects(ctx, &storage.Query{
		Prefix: prefix,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		c.deleteFile(ctx, attrs.Name)
	}

	return nil
}

// listFilenames - returns iterator to iterate over all matching objects
func (c *Client) listFilenames(ctx context.Context, prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var names []string

	// Get iterator for matching objects
	if !strings.HasPrefix(prefix, c.uploadPath) {
		prefix = c.uploadPath + prefix
	}
	it := c.cl.Bucket(c.bucketName).Objects(ctx, &storage.Query{
		Prefix: prefix,
	})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return names, err
		}
		names = append(names, attrs.Name)
	}

	return names, nil
}

// donwloadFirstFileByPrefix - downloads the first prefix matching file
func (c *Client) donwloadFirstFileByPrefix(ctx context.Context, prefix string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Get iterator for matching objects and return the first one
	if !strings.HasPrefix(prefix, c.uploadPath) {
		prefix = c.uploadPath + prefix
	}
	it := c.cl.Bucket(c.bucketName).Objects(ctx, &storage.Query{
		Prefix: prefix,
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []byte{}, "", err
		}

		file, err := c.downloadFile(ctx, attrs.Name)
		if err == nil {
			return file, attrs.Name, err
		}
	}
	return []byte{}, "", errors.New("not found")
}

// DetectMimeType - detects the mime type of a byte slice
func DetectMimeType(bytes []byte) *mimetype.MIME {
	return mimetype.Detect(bytes)
}

// DetectMimeTypeExtension - detects the mime type extension of a byte slice
func DetectMimeTypeExtension(bytes []byte) string {
	return DetectMimeType(bytes).Extension()
}
