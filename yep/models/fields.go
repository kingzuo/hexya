// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/npiganeau/yep/yep/orm"
	"github.com/npiganeau/yep/yep/tools"
)

var fieldsCache = &_fieldsCache{
	cache:                make(map[fieldRef]*fieldInfo),
	cacheByJSON:          make(map[fieldRef]*fieldInfo),
	computedFields:       make(map[string][]*fieldInfo),
	computedStoredFields: make(map[string][]*fieldInfo),
	dependencyMap:        make(map[fieldRef][]computeData),
}

/*
field is the key to find a field in the fieldsCache
*/
type fieldRef struct {
	modelName string
	name      string
}

// ConvertToName converts the given field ref to a fieldRef
// of type [modelName, fieldName].
func (fr *fieldRef) ConvertToName() {
	fi, ok := fieldsCache.get(*fr)
	if !ok {
		panic(fmt.Errorf("unknown fieldRef `%s`", *fr))
	}
	fr.name = fi.name
}

/*
computeData holds data to recompute another field.
- modelName is the name of the model to recompute
- compute is the name of the function to call on modelName
- path is the search string that will be used to find records to update.
The path should take an ID as argument (e.g. path = "Profile__BestPost").
*/
type computeData struct {
	modelName string
	compute   string
	path      string
}

/*
fieldInfo holds the meta information about a field
*/
type fieldInfo struct {
	modelName   string
	name        string
	column      string
	json        string
	description string
	help        string
	computed    bool
	stored      bool
	compute     string
	depends     []string
	html        bool
	fieldType   tools.FieldType
}

// fieldsCache is the fieldInfo collection
type _fieldsCache struct {
	sync.RWMutex
	cache                map[fieldRef]*fieldInfo
	cacheByJSON          map[fieldRef]*fieldInfo
	computedFields       map[string][]*fieldInfo
	computedStoredFields map[string][]*fieldInfo
	dependencyMap        map[fieldRef][]computeData
	done                 bool
}

/*
get returns the fieldInfo of the given field.
field can be of type [modelName, json_name] or [modelName, fieldName].
*/
func (fc *_fieldsCache) get(ref fieldRef) (fi *fieldInfo, ok bool) {
	fi, ok = fc.cache[ref]
	if !ok {
		fi, ok = fc.cacheByJSON[ref]
	}
	return
}

/*
getComputedFields returns the slice of fieldInfo of the computed, but not
stored fields of the given modelName.
If fields are given, return only fieldInfo in the list
*/
func (fc *_fieldsCache) getComputedFields(modelName string, fields ...string) (fil []*fieldInfo, ok bool) {
	fInfos, ok := fc.computedFields[modelName]
	if len(fields) > 0 {
		for _, f := range fields {
			for _, fInfo := range fInfos {
				if fInfo.name == tools.ConvertMethodName(f) {
					fil = append(fil, fInfo)
					continue
				}
			}
		}
	} else {
		fil = fInfos
	}
	return
}

/*
getComputedStoredFields returns the slice of fieldInfo of the computed and stored
fields of the given modelName.
*/
func (fc *_fieldsCache) getComputedStoredFields(modelName string) (fil []*fieldInfo, ok bool) {
	fil, ok = fc.computedStoredFields[modelName]
	return
}

/*
getDependentFields return the fields that must be recomputed when ref is modified.
*/
func (fc *_fieldsCache) getDependentFields(ref fieldRef) (target []computeData, ok bool) {
	target, ok = fc.dependencyMap[ref]
	return
}

/*
set adds the given fieldInfo to the fieldsCache.
ref must be of type [modelName, fieldName]
*/
func (fc *_fieldsCache) set(ref fieldRef, fInfo *fieldInfo) {
	fc.cache[ref] = fInfo
	colRef := fieldRef{modelName: ref.modelName, name: fInfo.json}
	fc.cacheByJSON[colRef] = fInfo
	if fInfo.computed {
		if fInfo.stored {
			fc.computedStoredFields[fInfo.modelName] = append(fc.computedStoredFields[fInfo.modelName], fInfo)
		} else {
			fc.computedFields[fInfo.modelName] = append(fc.computedFields[fInfo.modelName], fInfo)
		}
	}
}

/*
setDependency adds a dependency in the dependencyMap.
target field depends on ref field, i.e. when ref field is modified,
target field must be recomputed.
*/
func (fc *_fieldsCache) setDependency(ref fieldRef, target computeData) {
	fc.dependencyMap[ref] = append(fc.dependencyMap[ref], target)
}

/*
registerModelFields populates the fieldsCache with the given structPtr fields
*/
func registerModelFields(name string, structPtr interface{}) {
	var (
		attrs map[string]bool
		tags  map[string]string
	)

	val := reflect.ValueOf(structPtr)
	ind := reflect.Indirect(val)

	if val.Kind() != reflect.Ptr || ind.Kind() != reflect.Struct {
		panic(fmt.Errorf("<models.registerModelFields> cannot use non-ptr model struct `%s`", getModelName(structPtr)))
	}

	for i := 0; i < ind.NumField(); i++ {
		sf := ind.Type().Field(i)
		if sf.PkgPath != "" {
			// Skip private fields
			continue
		}
		parseStructTag(sf.Tag.Get(defaultStructTagName), &attrs, &tags)
		desc, ok := tags["string"]
		if !ok {
			desc = sf.Name
		}
		computeName, computed := tags["compute"]
		_, stored := attrs["store"]
		var depends []string
		if depTag, ok := tags["depends"]; ok {
			depends = strings.Split(depTag, defaultDependsTagDelim)
		}
		_, html := attrs["html"]
		ref := fieldRef{name: sf.Name, modelName: name}
		col := getFieldColumn(ref.modelName, ref.name)
		json, ok := tags["json"]
		if !ok {
			if col != "" {
				json = col
			} else {
				json = tools.SnakeCaseString(sf.Name)
			}
		}
		fInfo := fieldInfo{
			name:        sf.Name,
			column:      col,
			json:        json,
			modelName:   name,
			compute:     computeName,
			computed:    computed,
			stored:      stored,
			depends:     depends,
			description: desc,
			help:        tags["help"],
			html:        html,
		}
		fieldsCache.set(ref, &fInfo)
		fInfo.fieldType = getFieldType(ref)
	}
}

/*
processDepends populates the dependsMap of the fieldsCache from the depends strings of
each fieldInfo instance.
*/
func processDepends() {
	for targetField, fInfo := range fieldsCache.cache {
		var (
			refName string
		)
		for _, depString := range fInfo.depends {
			if depString != "" {
				tokens := strings.Split(depString, orm.ExprSep)
				refName = tokens[len(tokens)-1]
				refModelName := getRelatedModelName(targetField.modelName, depString)
				refField := fieldRef{
					modelName: refModelName,
					name:      refName,
				}
				path := strings.Join(tokens[:len(tokens)-1], orm.ExprSep)
				targetComputeData := computeData{
					modelName: fInfo.modelName,
					compute:   fInfo.compute,
					path:      path,
				}
				fieldsCache.setDependency(refField, targetComputeData)
			}
		}
	}
}

/*
getRelatedModelName returns the model name of the field given by path calculated from the origin model.
path is a query string as used in RecordSet.Filter()
*/
func getRelatedModelName(origin, path string) string {
	qs := orm.NewOrm().QueryTable(origin)
	modelName, _ := qs.TargetModelField(path)
	return modelName
}
