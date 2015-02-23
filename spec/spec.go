package spec

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/casualjim/go-swagger/assets"
	"github.com/casualjim/go-swagger/util"
)

const (
	// SwaggerSchemaURL the url for the swagger 2.0 schema to validate specs
	SwaggerSchemaURL = "http://swagger.io/v2/schema.json#"
	// JSONSchemaURL the url for the json schema schema
	JSONSchemaURL = "http://json-schema.org/draft-04/schema#"
)

//
// // Validate validates a spec document
// func Validate(doc *spec.Document, formats strfmt.Registry) *validate.Result {
// 	// TODO: add more validations beyond just jsonschema
// 	return newSchemaValidator(doc.Schema(), nil, "", formats).Validate(doc.Spec())
// }

// DocLoader represents a doc loader type
type DocLoader func(string) (json.RawMessage, error)

// JSONSpec loads a spec from a json document
func JSONSpec(path string) (*Document, error) {
	data, err := util.JSONDoc(path)
	if err != nil {
		return nil, err
	}
	// convert to json
	return New(json.RawMessage(data), "")
}

// YAMLSpec loads a swagger spec document
func YAMLSpec(path string) (*Document, error) {
	data, err := util.YAMLDoc(path)
	if err != nil {
		return nil, err
	}

	return New(data, "")
}

// MustLoadJSONSchemaDraft04 panics when Swagger20Schema returns an error
func MustLoadJSONSchemaDraft04() *Schema {
	d, e := JSONSchemaDraft04()
	if e != nil {
		panic(e)
	}
	return d
}

// JSONSchemaDraft04 loads the json schema document for json shema draft04
func JSONSchemaDraft04() (*Schema, error) {
	b, err := assets.Asset("jsonschema-draft-04.json")
	if err != nil {
		return nil, err
	}

	schema := new(Schema)
	if err := json.Unmarshal(b, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

// MustLoadSwagger20Schema panics when Swagger20Schema returns an error
func MustLoadSwagger20Schema() *Schema {
	d, e := Swagger20Schema()
	if e != nil {
		panic(e)
	}
	return d
}

// Swagger20Schema loads the swagger 2.0 schema from the embedded asses
func Swagger20Schema() (*Schema, error) {

	b, err := assets.Asset("2.0/schema.json")
	if err != nil {
		return nil, err
	}

	schema := new(Schema)
	if err := json.Unmarshal(b, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

// Document represents a swagger spec document
type Document struct {
	specAnalyzer
	spec *Swagger
	raw  json.RawMessage
}

var swaggerSchema *Schema
var jsonSchema *Schema

func init() {
	jsonSchema = MustLoadJSONSchemaDraft04()
	swaggerSchema = MustLoadSwagger20Schema()
}

// Load loads a new spec document
func Load(path string) (*Document, error) {
	specURL, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(specURL.Path)
	if ext == ".yaml" || ext == ".yml" {
		return YAMLSpec(path)
	}

	return JSONSpec(path)
}

// New creates a new shema document
func New(data json.RawMessage, version string) (*Document, error) {
	if version == "" {
		version = "2.0"
	}
	if version != "2.0" {
		return nil, fmt.Errorf("spec version %q is not supported", version)
	}

	spec := new(Swagger)
	if err := json.Unmarshal(data, spec); err != nil {
		return nil, err
	}

	d := &Document{
		specAnalyzer: specAnalyzer{
			spec:        spec,
			consumes:    make(map[string]struct{}),
			produces:    make(map[string]struct{}),
			authSchemes: make(map[string]struct{}),
			operations:  make(map[string]map[string]*Operation),
		},
		spec: spec,
		raw:  data,
	}
	d.initialize()
	return d, nil
}

// ExpandSpec expands the ref fields in the spec document
func (d *Document) ExpandSpec() error {
	// TODO: use a copy of the spec doc to expand instead
	// things requiring an expanded spec should first get the
	// expanded version of the document
	if err := expandSpec(d.spec); err != nil {
		return err
	}
	return nil
}

// BasePath the base path for this spec
func (d *Document) BasePath() string {
	return d.spec.BasePath
}

// Version returns the version of this spec
func (d *Document) Version() string {
	return d.spec.Swagger
}

// Schema returns the swagger 2.0 schema
func (d *Document) Schema() *Schema {
	return swaggerSchema
}

// Spec returns the swagger spec object model
func (d *Document) Spec() *Swagger {
	return d.spec
}

// Host returns the host for the API
func (d *Document) Host() string {
	return d.spec.Host
}

// Raw returns the raw swagger spec as json bytes
func (d *Document) Raw() json.RawMessage {
	return d.raw
}