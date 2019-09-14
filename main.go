package main

import (
	"github.com/gin-gonic/gin"
	"database/sql"
	"os"
	"log"
	"fmt"
	"net/http"
	"strconv"
	_ "github.com/lib/pq"
)

var db *sql.DB
var customers =[] Customer{}

type Customer struct{
	ID int `json:"id"`
    Name string `json:"name"`
	Email string `json:"email"`
	Status string `json:"status"`
}

func connectDB() (*sql.DB, error) {
	db,err:=sql.Open("postgres",os.Getenv("DATABASE_URL"))
	return db,err
}

func createCustomerHandler(c *gin.Context){
	db, err:=connectDB()
	if err!= nil {
		log.Println("Connect to database error",err)
	}
	defer db.Close()
	var cust Customer
	err= c.ShouldBindJSON(&cust)
	if err!=nil{
		c.JSON(http.StatusBadRequest,err.Error())
	}
	row:= db.QueryRow("INSERT INTO customers (name,email,status) values($1,$2,$3) RETURNING id",cust.Name,cust.Email,cust.Status)

	var id int
	err = row.Scan(&id)
	if err !=nil{
		fmt.Println("can't scan if",err)
		return
	}
	cust.ID=id
	fmt.Println("insert customer success id",id)
	c.JSON(http.StatusCreated,cust)
}

func getCustByIDHandler(c *gin.Context){
	db, err:=connectDB()
	if err!= nil {
		log.Println("Connect to database error",err)
	}
	defer db.Close()
	cid:=c.Param("id")
	stmt,err:=db.Prepare("SELECT id,name,email,status FROM customers where id=$1")
	if err !=nil{
		log.Fatal("can't prepare query one row statement",err)
	}
	rowID,err:=strconv.Atoi(cid)
	row :=stmt.QueryRow(rowID)
	var id int
	var name,email,status string
	err= row.Scan(&id,&name,&email,&status)
	if err!=nil{
		log.Fatal("can't Scan row into variables ",err)
	}
	cust:=Customer{}
	cust.ID=id
	cust.Name=name
	cust.Email=email
	cust.Status=status
	c.JSON(http.StatusOK,cust)
}

func getAllCustHandler(c *gin.Context){
	db,err:=connectDB()
	if err!= nil {
		log.Println("Connect to database error",err)
	}
	defer db.Close()
	stmt,err := db.Prepare("SELECT id,name,email,status FROM customers")
	if err != nil {
		log.Fatal("can't prepare query all todos statment", err)
	}
	rows,err:= stmt.Query()
	if err != nil {
		log.Fatal("can't query all customer", err)
	}
	customers:=[] Customer{}
	for rows.Next(){
		var id int
		var name,email,status string
		err:=rows.Scan(&id,&name,&email,&status)
		if err != nil {
			log.Fatal("can't Scan row into variable", err)
		}
		cust:=Customer{}
		cust.ID=id
		cust.Name=name
		cust.Email=email
		cust.Status=status
		customers=append(customers,cust)
	}
	fmt.Println("query all todos success")
	c.JSON(http.StatusOK,customers)
}

func updateCustByIDHandler(c *gin.Context){
	db,err:=connectDB()
	if err!= nil {
		log.Println("Connect to database error",err)
	}
	defer db.Close()
	cid:=c.Param("id")
	icid,_:=strconv.Atoi(cid)
	var cust Customer
	err= c.ShouldBindJSON(&cust)
	if (err!=nil) || (icid !=cust.ID) {
		c.JSON(http.StatusBadRequest,err.Error())
	}
	
	stmt, err := db.Prepare("UPDATE customers SET name=$2,email=$3,status=$4 WHERE id=$1;")
	if err !=nil{
		log.Fatal("can't prepare statment update", err)
	}
	_,err =stmt.Exec(cust.ID,cust.Name,cust.Email,cust.Status)
	if err!=nil{
		log.Fatal("error execute update ", err)
	}
	fmt.Println("update customers success")
	c.JSON(http.StatusOK,cust)
}

func deleteCustByIDHandler(c *gin.Context){
	db,err:=connectDB()
	if err!= nil {
		log.Println("Connect to database error",err)
	}
	defer db.Close()
	cid:=c.Param("id")
	icid,_:=strconv.Atoi(cid)
	stmt, err := db.Prepare("DELETE FROM customers WHERE id=$1;")
	if err !=nil{
		log.Fatal("can't prepare statment Delete ", err)
	}
	_,err =stmt.Exec(icid)
	if err!=nil{
		log.Fatal("error execute delete ", err)
	}
	fmt.Println("delete customers success")
	c.JSON(http.StatusOK, map[string]string{"message":"customer deleted"})
}

func authMiddleware(c *gin.Context){
	token:=c.GetHeader("Authorization")
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized,gin.H{"error":"unauthorized"})
		c.Abort()
		return
	}
	c.Next()
	fmt.Println("Authen pass")
}

func createTB(){
	db,err:= connectDB()
	if err!=nil{
		log.Fatal("Connect to database error",err)
	}
	defer db.Close()
	fmt.Println("ok")
	createTb:=`
	CREATE TABLE IF NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT 
	);
	`
	_,err=db.Exec(createTb)
	if err !=nil{
		log.Fatal("can't create table",err)
	}
	fmt.Println("create table success")

}

func main(){
	createTB()
	r:=gin.Default()
	r.Use(authMiddleware)

	r.POST("/customers",createCustomerHandler)
	r.GET("/customers/:id",getCustByIDHandler)
	r.GET("/customers",getAllCustHandler)
	r.PUT("/customers/:id",updateCustByIDHandler)
	r.DELETE("/customers/:id",deleteCustByIDHandler)

	r.Run(":2019")
}