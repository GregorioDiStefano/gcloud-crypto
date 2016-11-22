package main

// Bucket is an interface that specifies all the basic functionalities
// a cloud storage service must impplement.
type Bucket interface {
	// Delete file from bucket
	Delete(name string) error
	// Upload file to bucket
	Upload(string, string, []byte) error
	// Download file from bucket
	Download(name string) (string, error)
	// List files in the bucket
	List() ([]string, error)
	// Move the file
	Move(string, string) error
}
