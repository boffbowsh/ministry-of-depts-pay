package models

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
  "net/http"
  "io/ioutil"
	"encoding/json"
	"os"

	"github.com/astaxie/beego/orm"
)

type Department struct {
	Id        int64  `orm:"auto"`
	Name      string `orm:"size(128)"`
	PaymentId string `orm:"size(128)"`
	Status    string `orm:"size(128)"`
	Reference string `orm:"size(128)"`
}

type CreateResponse struct {
	Links []struct {
		Rel string `json:"rel"`
		Method string `json:"method"`
		Href string `json:"href"`
	} `json:"links"`
	Amount int `json:"amount"`
	Status string `json:"status"`
	Description string `json:"description"`
	Reference string `json:"reference"`
	PaymentId string `json:"payment_id"`
	ReturnURL string `json:"return_url"`
}

func init() {
	orm.RegisterModel(new(Department))
}

// AddDepartment insert a new Department into database and returns
// last inserted Id on success.
func AddDepartment(m *Department) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

func GetRedirectUrl(m *Department) (string) {
		url := "https://publicapi-integration-1.pymnt.uk//v1/payments"

		payload := &struct{
			ReturnUrl string `json:"return_url"`
			Amount int64 `json:"amount"`
			Reference string `json:"reference"`
			Description string `json:"description"`
			AccountId string `json:"account_id"`
		} {
			fmt.Sprintf("http://localhost:8080/departments/%d", m.Id),
			100000,
			m.Reference,
			fmt.Sprintf("Setup fee for %s", m.Name),
			"",
		}

		data, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", url, bytes.NewBufferString(string(data)))

    req.Header.Add("accept", "application/json")
    req.Header.Add("authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")))
    req.Header.Add("content-type", "application/json")

    res, _ := http.DefaultClient.Do(req)

    defer res.Body.Close()
    body, _ := ioutil.ReadAll(res.Body)

		response := &CreateResponse{}
		json.Unmarshal(body, &response)

		o := orm.NewOrm()
		v := Department{Id: m.Id}
		o.Read(&v)
		v.PaymentId = response.PaymentId
		o.Update(&v, "PaymentID")

		for _, v := range response.Links {
			if v.Rel == "next_url" {
				return v.Href
			}
		}

		panic(string(body))
}

func CheckPaymentStatus(m *Department) (string) {
	url := fmt.Sprintf("https://publicapi-integration-1.pymnt.uk//v1/payments/%s", m.PaymentId)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_KEY")))

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

	response := &struct{
		Status string `json:"status"`
	}{}

	json.Unmarshal(body, &response)

	o := orm.NewOrm()
	v := Department{Id: m.Id}
	o.Read(&v)
	v.Status = response.Status
	o.Update(&v, "Status")

	return response.Status
}

// GetDepartmentById retrieves Department by Id. Returns error if
// Id doesn't exist
func GetDepartmentById(id int64) (v *Department, err error) {
	o := orm.NewOrm()
	v = &Department{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllDepartment retrieves all Department matches certain condition. Returns empty list if
// no records exist
func GetAllDepartment(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	o := orm.NewOrm()
	qs := o.QueryTable(new(Department))
	// query k=v
	for k, v := range query {
		// rewrite dot-notation to Object__Attribute
		k = strings.Replace(k, ".", "__", -1)
		qs = qs.Filter(k, v)
	}
	// order by:
	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			// 1) for each sort field, there is an associated order
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			// 2) there is exactly one order, all the sorted fields will be sorted by this order
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]")
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1")
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields")
		}
	}

	var l []Department
	qs = qs.OrderBy(sortFields...)
	if _, err := qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(fname).Interface()
				}
				ml = append(ml, m)
			}
		}
		return ml, nil
	}
	return nil, err
}

// UpdateDepartment updates Department by Id and returns error if
// the record to be updated doesn't exist
func UpdateDepartmentById(m *Department) (err error) {
	o := orm.NewOrm()
	v := Department{Id: m.Id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteDepartment deletes Department by Id and returns error if
// the record to be deleted doesn't exist
func DeleteDepartment(id int64) (err error) {
	o := orm.NewOrm()
	v := Department{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Department{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}
