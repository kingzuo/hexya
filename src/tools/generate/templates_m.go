// Copyright 2019 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package generate

import "text/template"

var poolInterfacesTemplate = template.Must(template.New("").Parse(`
// This file is autogenerated by hexya-generate
// DO NOT MODIFY THIS FILE - ANY CHANGES WILL BE OVERWRITTEN

package {{ .InterfacesPackageName }}

import (
    "github.com/hexya-erp/pool/{{ .QueryPackageName }}"
	{{ range .Deps }} 	"{{ . }}"
	{{ end }}
)

// {{ .Name }}Set is an autogenerated type to handle {{ .Name }} objects.
type {{ .Name }}Set interface {
	models.RecordSet
	// {{ .Name }}SetHexyaFunc is a dummy function to uniquely match interfaces.
	{{ .Name }}SetHexyaFunc()
	{{- range .Fields }}
	// {{ .Name }} is a getter for the value of the "{{ .Name }}" field of the first
	// record in this RecordSet. It returns the Go zero value if the RecordSet is empty.
	{{ .Name }}() {{ .IType }}
	// Set{{ .Name }} is a setter for the value of the "{{ .Name }}" field of this
	// RecordSet. All Records of this RecordSet will be updated. Each call to this
	// method makes an update query in the database.
	//
	// Set{{ .Name }} panics if the RecordSet is empty.
	Set{{ .Name }}(value {{ .IType }})
	{{- end }}
	{{- range .AllMethods }}
	{{ .Doc }}
	{{ .Name }}({{ .IParamsTypes }}) ({{ .IReturnString }})
	{{- end }}
	// Super returns a RecordSet with a modified callstack so that call to the current
	// method will execute the next method layer.
	//
	// This method is meant to be used inside a method layer function to call its parent,
	// such as:
	//
	//    func (rs h.MyRecordSet) MyMethod() string {
	//        res := rs.Super().MyMethod()
	//        res += " ok!"
	//        return res
	//    }
	//
	// Calls to a different method than the current method will call its next layer only
	// if the current method has been called from a layer of the other method. Otherwise,
	// it will be the same as calling the other method directly.
	Super() {{ .Name }}Set
	// ModelData returns a new {{ .Name }}Data object populated with the values
	// of the given FieldMap. 
	ModelData(fMap models.FieldMap) {{ .Name }}Data
	// Records returns a slice with all the records of this RecordSet, as singleton RecordSets
	Records() []{{ .Name }}Set
	// First returns the values of the first Record of the RecordSet as a pointer to a {{ .Name }}Data.
	//
	// If this RecordSet is empty, it returns an empty {{ .Name }}Data.
	First() {{ .Name }}Data
	// All returns the values of all Records of the RecordCollection as a slice of {{ .Name }}Data pointers.
	All() []{{ .Name }}Data
}

// {{ .Name }}Data is used to hold values of an {{ .Name }} object instance
// when creating or updating a {{ .Name }}Set.
type {{ .Name }}Data interface {
	// Underlying returns the object converted to a FieldMap.
	Underlying() models.FieldMap
	// Get returns the value of the given field.
	// The second returned value is true if the value exists.
	//
	// The field can be either its name or is JSON name.
	Get(field string) (interface{}, bool)
	// Set sets the given field with the given value.
	// If the field already exists, then it is updated with value.
	// Otherwise, a new entry is inserted.
	//
	// It returns the given {{ .Name }}Data so that calls can be chained
	Set(field string, value interface{}) {{ .Name }}Data
	// Unset removes the value of the given field if it exists.
	//
	// It returns the given ModelData so that calls can be chained
	Unset(field string) {{ .Name }}Data
	// Copy returns a copy of this {{ .Name }}Data	
	Copy() {{ .Name }}Data
	// Keys returns the {{ .Name }}Data keys as a slice of strings
	Keys() (res []string)
	// OrderedKeys returns the keys of this {{ .Name }}Data ordered.
	//
	// This has the convenient side effect of having shorter paths come before longer paths,
	// which is particularly useful when creating or updating related records.
	OrderedKeys() []string
	// FieldNames returns the {{ .Name }}Data keys as a slice of FieldNamer.
	FieldNames() (res []models.FieldNamer)
	{{- range .Fields }}
	// {{ .Name }} returns the value of the {{ .Name }} field.
	// If this {{ .Name }} is not set in this {{ $.Name }}Data, then
	// the Go zero value for the type is returned.
	{{ .Name }}() {{ .IType }}
	// Has{{ .Name }} returns true if {{ .Name }} is set in this {{ $.Name }}Data
	Has{{ .Name }}() bool
	// Set{{ .Name }} sets the {{ .Name }} field with the given value.
	// It returns this {{ $.Name }}Data so that calls can be chained.
	Set{{ .Name }}(value {{ .IType }}) {{ $.Name }}Data
	// Unset{{ .Name }} removes the value of the {{ .Name }} field if it exists.
	// It returns this {{ $.Name }}Data so that calls can be chained.
	Unset{{ .Name }}() {{ $.Name }}Data
	{{- end }}
}

// A {{ .Name }}GroupAggregateRow holds a row of results of a query with a group by clause
type {{ .Name }}GroupAggregateRow interface {
	// Values() returns the values of the actual query
	Values() {{ .Name }}Data
	// Count is the number of lines aggregated into this one
	Count() int
	// Condition can be used to query the aggregated rows separately if needed
	Condition() {{ $.QueryPackageName }}.{{ .Name }}Condition
}

`))
