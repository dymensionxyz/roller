package structs

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/cosmos/cosmos-sdk/types"
)

func InitializeMetadata(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			if field.String() == "" {
				field.SetString("")
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if field.Int() == 0 {
				field.SetInt(0)
			}
		case reflect.Float32, reflect.Float64:
			if field.Float() == 0 {
				field.SetFloat(0)
			}
		case reflect.Bool:
			if !field.Bool() {
				field.SetBool(false)
			}
		case reflect.Slice:
			if field.IsNil() {
				field.Set(reflect.MakeSlice(field.Type(), 0, 0))
			}
		case reflect.Map:
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}
		case reflect.Struct:
			InitializeMetadata(field)
		case reflect.Ptr:
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			InitializeMetadata(field.Elem())
		}
	}

	// Special handling for cosmossdkmath.Int
	if v.Type().Name() == "Metadata" {
		if gasPriceField := v.FieldByName("GasPrice"); gasPriceField.IsValid() {
			if gasPriceField.IsNil() {
				gasPriceField.Set(reflect.ValueOf(types.NewInt(0)))
			}
		}
	}
}

func ExportStructToFile(data interface{}, filename string) error {
	// Initialize the struct with default values
	InitializeMetadata(reflect.ValueOf(data))

	// Marshal the struct to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	// Write to file
	err = os.WriteFile(filename, jsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
