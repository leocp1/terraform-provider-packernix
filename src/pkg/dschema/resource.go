package dschema

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// An interface that generalizes schema.ResourceDiff and schema.ResourceData.
// Set is only supported on computed fields
type resource interface {
	Get(string) interface{}
	GetChange(string) (interface{}, interface{})
	GetOk(string) (interface{}, bool)
	Id() string
	Set(string, interface{}) error
}

type rdiffpassthru struct {
	Rd *schema.ResourceDiff
}

func (rd *rdiffpassthru) Get(key string) interface{} {
	return rd.Rd.Get(key)
}
func (rd *rdiffpassthru) GetChange(key string) (interface{}, interface{}) {
	return rd.Rd.GetChange(key)
}
func (rd *rdiffpassthru) GetOk(key string) (interface{}, bool) {
	return rd.Rd.GetOk(key)
}
func (rd *rdiffpassthru) Id() string {
	return rd.Rd.Id()
}
func (rd *rdiffpassthru) Set(key string, value interface{}) error {
	return rd.Rd.SetNew(key, value)
}
func ResourceDiffAdapter(rd *schema.ResourceDiff) resource {
	return &rdiffpassthru{Rd: rd}
}
