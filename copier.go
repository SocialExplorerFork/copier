package copier

import "reflect"

const (
	ValidFieldName = "Valid"
)

func CopyOnlyValid(copy_to interface{}, copy_from interface{}) (err error) {
	return copy_internal(copy_to, copy_from, true)
}

func Copy(copy_to interface{}, copy_from interface{}) (err error) {
	return copy_internal(copy_to, copy_from, false)
}

func copy_internal(copy_to interface{}, copy_from interface{}, only_valid bool) (err error) {
	var (
		is_slice        bool
		from_typ        reflect.Type
		is_from_typ_ptr bool
		to_typ          reflect.Type
		is_to_typ_ptr   bool
		elem_amount     int
	)

	from := reflect.ValueOf(copy_from)
	to := reflect.ValueOf(copy_to)
	from_elem := reflect.Indirect(from)
	to_elem := reflect.Indirect(to)

	if to_elem.Kind() == reflect.Slice {
		is_slice = true
		if from_elem.Kind() == reflect.Slice {
			from_typ = from_elem.Type().Elem()
			if from_typ.Kind() == reflect.Ptr {
				from_typ = from_typ.Elem()
				is_from_typ_ptr = true
			}
			elem_amount = from_elem.Len()
		} else {
			from_typ = from_elem.Type()
			elem_amount = 1
		}

		to_typ = to_elem.Type().Elem()
		if to_typ.Kind() == reflect.Ptr {
			to_typ = to_typ.Elem()
			is_to_typ_ptr = true
		}

	} else {
		from_typ = from_elem.Type()
		to_typ = to_elem.Type()
		elem_amount = 1
	}

	for e := 0; e < elem_amount; e++ {
		var dest, source reflect.Value
		if is_slice {
			if from_elem.Kind() == reflect.Slice {
				source = from_elem.Index(e)
				if is_from_typ_ptr {
					source = source.Elem()
				}
			} else {
				source = from_elem
			}
		} else {
			source = from_elem
		}

		if is_slice {
			dest = reflect.New(to_typ).Elem()
		} else {
			dest = to_elem
		}

		for i := 0; i < from_typ.NumField(); i++ {
			field := from_typ.Field(i)
			if !field.Anonymous {

				name := field.Name
				field_type := field.Type

				if only_valid && field_type.Kind() == reflect.Struct {
					_, found := field_type.FieldByName(ValidFieldName)

					if found && source.IsValid() && source.FieldByName(name).IsValid() {
						valid_field_v := source.FieldByName(name).FieldByName(ValidFieldName)
						if valid_field_v.Kind() == reflect.Bool {
							isValid := valid_field_v.Bool()
							if !isValid {
								continue
							}
						}
					}
				}

				from_field := source.FieldByName(name)
				to_field := dest.FieldByName(name)
				to_method := dest.Addr().MethodByName(name)
				if from_field.IsValid() && to_field.IsValid() {
					to_field.Set(from_field)
				}

				if from_field.IsValid() && to_method.IsValid() {
					to_method.Call([]reflect.Value{from_field})
				}
			}
		}

		for i := 0; i < dest.NumField(); i++ {
			field := to_typ.Field(i)
			if !field.Anonymous {
				name := field.Name
				field_type := field.Type

				if only_valid && field_type.Kind() == reflect.Struct {
					_, found := field_type.FieldByName(ValidFieldName)
					if found && source.IsValid() && source.FieldByName(name).IsValid() {
						valid_field_v := source.FieldByName(name).FieldByName(ValidFieldName)
						if valid_field_v.Kind() == reflect.Bool {

							isValid := valid_field_v.Bool()
							if !isValid {
								continue
							}
						}
					}
				}

				from_method := source.Addr().MethodByName(name)
				to_field := dest.FieldByName(name)

				if from_method.IsValid() && to_field.IsValid() {
					values := from_method.Call([]reflect.Value{})
					if len(values) >= 1 {
						to_field.Set(values[0])
					}
				}
			}
		}

		if is_slice {
			if is_to_typ_ptr {
				to_elem.Set(reflect.Append(to_elem, dest.Addr()))
			} else {
				to_elem.Set(reflect.Append(to_elem, dest))
			}

		}
	}
	return
}
