package main

import (
	"database/sql"
	_ "database/sql" // add this
	"fmt"
	"log"
	"reflect"

	_ "github.com/lib/pq" // add this

	"net/http"

	"github.com/gin-gonic/gin"
)

func RowsToStructs(rows *sql.Rows, dest interface{}) error {

	// 1. Create a slice of structs from the passed struct type of model
	//
	// Not needed, the caller passes pointer to destination slice.
	// Elem() dereferences the pointer.
	//
	// If you do need to create the slice in this function
	// instead of using the argument, then use
	// destv := reflect.MakeSlice(reflect.TypeOf(model).

	destv := reflect.ValueOf(dest).Elem()

	// Allocate argument slice once before the loop.

	args := make([]interface{}, destv.Type().Elem().NumField())

	// 2. Loop through each row

	for rows.Next() {

		// 3. Create a struct of passed mode interface{} type
		rowp := reflect.New(destv.Type().Elem())
		rowv := rowp.Elem()

		// 4. Scan the row results to a slice of interface{}
		// 5. Set the field values of struct created in step 3 using the slice in step 4
		//
		// Scan directly to the struct fields so the database
		// package handles the conversion from database
		// types to a Go types.
		//
		// The slice args is filled with pointers to struct fields.

		for i := 0; i < rowv.NumField(); i++ {
			args[i] = rowv.Field(i).Addr().Interface()
		}

		if err := rows.Scan(args...); err != nil {
			return err
		}

		// 6. Add the struct created in step 3 to slice created in step 1

		destv.Set(reflect.Append(destv, rowv))

	}
	return nil
}

func getAllHandler(c *gin.Context, db *sql.DB) {
	var students []struct {
		Student_id       uint64
		Student_name     string
		Student_age      uint64
		Student_address  string
		Student_phone_no string
	}
	rows, err := db.Query("SELECT * FROM student")
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
		c.JSON(http.StatusBadRequest, "An error occured")
	}
	RowsToStructs(rows, &students)

	if students == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Data empty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": students})
}

func getHandler(c *gin.Context, db *sql.DB) {
	var students []struct {
		Student_id       uint64
		Student_name     string
		Student_age      uint64
		Student_address  string
		Student_phone_no string
	}
	student_id := c.Params.ByName("studentId")
	rows, err := db.Query("SELECT * FROM student WHERE student_id=$1", student_id)
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
		c.JSON(http.StatusBadRequest, "An error occured")
	}
	RowsToStructs(rows, &students)

	if students == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Data empty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": students})
}

func postHandler(c *gin.Context, db *sql.DB) {
	var newStudent struct {
		Student_id       uint64 `json:"student_id" binding:"required"`
		Student_name     string `json:"student_name" binding:"required"`
		Student_age      uint64 `json:"student_age" binding:"required"`
		Student_address  string `json:"student_address" binding:"required"`
		Student_phone_no string `json:"student_phone_no" binding:"required"`
	}

	if c.Bind(&newStudent) == nil {

		fmt.Printf("%v", newStudent)
		_, err := db.Exec("INSERT into student VALUES ($1, $2, $3, $4, $5)", newStudent.Student_id, newStudent.Student_name, newStudent.Student_age, newStudent.Student_address, newStudent.Student_phone_no)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "create success"})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"message": "error"})
}

func putHandler(c *gin.Context, db *sql.DB) {
	var students []struct {
		Student_id       uint64
		Student_name     string
		Student_age      uint64
		Student_address  string
		Student_phone_no string
	}
	student_id := c.Params.ByName("studentId")
	rows, err := db.Query("SELECT * FROM student WHERE student_id=$1", student_id)
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
		c.JSON(http.StatusBadRequest, "An error occured")
	}
	RowsToStructs(rows, &students)

	if students == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		return
	}

	var newStudent struct {
		Student_name     string `json:"student_name" binding:"required"`
		Student_age      uint64 `json:"student_age" binding:"required"`
		Student_address  string `json:"student_address" binding:"required"`
		Student_phone_no string `json:"student_phone_no" binding:"required"`
	}

	if c.Bind(&newStudent) == nil {

		fmt.Printf("%v", newStudent)
		db.Exec("UPDATE student SET student_name=$1 WHERE student_id=$2", newStudent.Student_name, student_id)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "failed to update"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "update success"})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"message": "error"})
}

func deleteHandler(c *gin.Context, db *sql.DB) {
	var students []struct {
		Student_id       uint64
		Student_name     string
		Student_age      uint64
		Student_address  string
		Student_phone_no string
	}
	student_id := c.Params.ByName("studentId")
	rows, err := db.Query("SELECT * FROM student WHERE student_id=$1", student_id)
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
		c.JSON(http.StatusBadRequest, "An error occured")
	}
	RowsToStructs(rows, &students)

	if students == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Data not found"})
		return
	}

	_, errDelete := db.Exec("DELETE from student WHERE student_id=$1", student_id)

	if errDelete != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to delete", "detail": errDelete.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "delete success"})
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	gin.DisableConsoleColor()
	r := gin.Default()

	connStr := "postgresql://postgres:postgres@127.0.0.1/go_crud?sslmode=disable"
	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/student", func(c *gin.Context) {
		getAllHandler(c, db)
	})

	// Get student value
	r.GET("/student/:studentId", func(c *gin.Context) {
		getHandler(c, db)
	})

	r.POST("/student", func(c *gin.Context) {
		postHandler(c, db)
	})

	r.PUT("/student/:studentId", func(c *gin.Context) {
		putHandler(c, db)
	})

	r.DELETE("/student/:studentId", func(c *gin.Context) {
		deleteHandler(c, db)
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
