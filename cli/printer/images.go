package printer

import "github.com/hostinger/fireactions/api"

// Image is a Printable for api.Images
type Image struct {
	Images api.Images
}

var _ Printable = &Image{}

// Cols returns the columns for the Printable
func (i *Image) Cols() []string {
	cols := []string{
		"ID", "Name", "URL",
	}

	return cols
}

// ColsMap returns the columns map for the Printable
func (i *Image) ColsMap() map[string]string {
	cols := map[string]string{
		"ID": "ID", "Name": "Name", "URL": "URL",
	}

	return cols
}

// KV returns the key value for the Printable
func (i *Image) KV() []map[string]interface{} {
	kv := make([]map[string]interface{}, 0, len(i.Images))
	for _, image := range i.Images {
		kv = append(kv, map[string]interface{}{
			"ID": image.ID, "Name": image.Name, "URL": image.URL,
		})
	}

	return kv
}
