package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {

	lambda.Start(handleRequest)
	// handle()

}

//SupplierSQL for database
type SupplierSQL struct {
	ID                  int    `db:"origin"`
	Type                string `db:"origin_type"`
	BookingEngineStatus int    `db:"booking_engine_status"`
	SourceID            string `db:"source_id"`
	Name                string `db:"name"`
	Description         string `db:"description"`
	CreatedFirstName    string `db:"first_name"`
	CreatedLastName     string `db:"last_name"`
	CreatedDate         string `db:"created_datetime"`
	Authentication      int    `db:"authentication"`
	Config              string `db:"config"`
	Remarks             string `db:"remarks"`
	UserID              string `db:"created_by_id"`
}

//Supplier to json
type Supplier struct {
	ID                  int                    `json:"origin"`
	Type                string                 `json:"originType"`
	BookingEngineStatus int                    `json:"bookingEngineStatus"`
	SourceID            string                 `json:"sourceId"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Config              map[string]interface{} `json:"config"`
	CreatedFirstName    string                 `json:"firstName"`
	CreatedLastName     string                 `json:"lastName"`
	CreatedDate         string                 `json:"createdDatetime"`
	Authentication      int                    `json:"authentication"`
	Remarks             string                 `json:"remarks"`
}

func handleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// func handle() {

	user := os.Getenv("b2b_user")
	pass := os.Getenv("b2b_password")
	url := os.Getenv("b2b_url")
	schema := os.Getenv("b2b_schema")

	db, err := sqlx.Connect("mysql", user+":"+pass+"@tcp("+url+")/"+schema)
	if err != nil {
		panic(err)
	}
	rows, err := db.Queryx(`select wal.origin , wal.origin_type, wal.booking_engine_status ,du.first_name , du.last_name 
	, wal.source_id , wal.name , wal.description 
	, wal.created_by_id, wal.created_datetime
	,wal.authentication , wac.config, wac.remarks 
	from webservice_api_list wal, webservice_api_credential wac, domain_user du where wal.origin =  wac.origin and du.user_id = wal.created_by_id`)
	if err != nil {
		panic(err)
	}

	suppliers := []Supplier{}
	for rows.Next() {
		supplierSQL := SupplierSQL{}
		err = rows.StructScan(&supplierSQL)
		if err != nil {
			rows.Close()
			panic(err)
		}
		supplier := bookingDBToJSON(supplierSQL)

		var Attributes map[string]interface{}

		json.Unmarshal([]byte(supplierSQL.Config), &Attributes)
		supplier.Config = Attributes

		suppliers = append(suppliers, supplier)
	}
	defer db.Close()

	respBody, err := json.Marshal(suppliers)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("matata =>", string(respBody))

	retVal := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(respBody),
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}
	return retVal, nil

}

func bookingDBToJSON(b SupplierSQL) Supplier {
	p := Supplier{}
	p.ID = b.ID
	p.Type = b.Type
	p.BookingEngineStatus = b.BookingEngineStatus
	p.SourceID = b.SourceID
	p.Name = b.Name
	p.Description = b.Description
	p.CreatedFirstName = b.CreatedFirstName
	p.CreatedLastName = b.CreatedLastName
	p.CreatedDate = b.CreatedDate
	p.Authentication = b.Authentication
	p.Remarks = b.Remarks
	return p
}
