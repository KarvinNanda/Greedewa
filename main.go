package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/mail"
	_ "os"
	_ "path/filepath"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct{
	UserID int 
	Username string 
	Password string 
	Email string 
}

type Item struct{
	ItemID int 
	CategoriesID int 
	ItemName string
	ItemPrice int 
	ItemImage string
}

var dsn = "root:@tcp(127.0.0.1:3306)/greedewa?charset=utf8mb4&parseTime=True&loc=Local"
var db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{})

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("./*.html")

	r.Static("/css", "./css/")
	r.Static("/scss", "./scss/")
	r.Static("/js", "./js/")
	r.Static("/images","./images/")
	r.Static("/fonts","./fonts/")
	r.Static("/vendor","./vendor/")

	store := cookie.NewStore([]byte("session_id"))
	r.Use(sessions.Sessions("user",store))

	r.GET("/",index)
	r.GET("/sayur",sayur)
	r.GET("/buah",buah)
	r.GET("/hidroponik",hidro)
	r.GET("/eksklusif",eksklu)
	r.GET("/shop",shop)
	
	r.GET("/contact",contact)

	r.GET("/login",login)
	r.GET("/register",register)
	r.GET("/logout",logout)

	r.GET("/delete/:id",deleteProduct)
	r.GET("/pay/:id",payProduct)
	r.GET("/edit/:id",edit)
	r.GET("/create",create)

	r.POST("/login",doLogin)
	r.POST("/register",doRegister)
	r.POST("/edit",doEdit)
	
	r.Run(":1234")
}

func create(c *gin.Context){
	ses := sessions.Default(c)
	name := ses.Get("name")


	if name == nil{
		c.Redirect(http.StatusMovedPermanently,"/shop")
	}
	c.HTML(http.StatusOK,"item.html",gin.H{"":""})
}

func doEdit(c *gin.Context){
	var findItem Item
	

	id := c.PostForm("id")
	name := c.PostForm("name")
	price := c.PostForm("price")
	categories := c.PostForm("categories")
	file,_,_ := c.Request.FormFile("file")

	filebytes,_ := ioutil.ReadAll(file)

	p,_ := strconv.Atoi(price)

	cat,_ := strconv.Atoi(categories)

	if id == ""{
		findItem.ItemName = name
		findItem.ItemPrice = p
		findItem.CategoriesID = cat
		findItem.ItemImage = string(filebytes)
		db.Create(&findItem)
	} else {
		findItem.ItemName = name
		findItem.ItemPrice = p
		findItem.CategoriesID = cat
		findItem.ItemImage = string(filebytes)
		db.Where("item_id = ?",id).Updates(&findItem)
	}

	c.Redirect(http.StatusMovedPermanently,"/shop")
	
}

func deleteProduct(c *gin.Context){
	id := c.Param("id")
	var findItem Item
	db.Where("item_id = ?",id).Delete((&findItem))
	
	ses := sessions.Default(c)
	name := ses.Get("name")


	if name == nil{
		c.Redirect(http.StatusMovedPermanently,"/shop")
	}
	
	c.Redirect(http.StatusMovedPermanently,"/shop")
}

func payProduct(c *gin.Context){
	id := c.Param("id")
	var findItem Item
	var findUser User
	db.Where("item_id = ?",id).Find((&findItem))

	
	ses := sessions.Default(c)
	name := ses.Get("name")


	if name == nil{
		c.Redirect(http.StatusMovedPermanently,"/shop")
	} else {
		
		db.Where("username = ?",name).Find(&findUser)
	}

	

	c.HTML(http.StatusOK,"checkout.html",gin.H{
		"Item":findItem,
		"User":findUser,
		"Name":name,

	})
}

func edit(c *gin.Context){
	id := c.Param("id")
	var findItem Item
	db.Where("item_id = ?",id).Find((&findItem))
	
	ses := sessions.Default(c)
	name := ses.Get("name")


	if name == nil{
		c.Redirect(http.StatusMovedPermanently,"/shop")
	}


	c.HTML(http.StatusOK,"item.html",gin.H{"Item":findItem,})
}

func logout(c *gin.Context){
	ses := sessions.Default(c)
	ses.Clear()
	ses.Save()
	// item := GetItems(2)
	fmt.Println("lokout")

	c.Redirect(http.StatusMovedPermanently,"/")
	// c.HTML(http.StatusOK,"index.html",nil)
}

func doLogin(c *gin.Context){
	username := c.PostForm("username")
	pass := c.PostForm("pass")
	var user User
	var msg []string
	ses := sessions.Default(c)
	

	db.Where("username = ?",username).Find(&user)

	err := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(pass))
	fmt.Println(err)

	if err != nil {
		msg = append(msg, "user not found")
	}

	if len(msg) > 0{
		c.HTML(http.StatusOK,"login.html",gin.H{"msg":msg,})
	} else {
		// item := GetItems(2)
		ses.Set("name",user.Username)
		ses.Save()
		c.Redirect(http.StatusMovedPermanently,"/")
	}
}

func doRegister(c *gin.Context){
	email := c.PostForm("email")
	username := c.PostForm("username")
	pass := c.PostForm("pass")
	var msg []string
	var user User

	if len(email) == 0 {
		msg=append(msg, "email must filled")
	} else if _,err := mail.ParseAddress(email);err!=nil {
		msg=append(msg, "email not valid")
	}

	if len(username) == 0 {
			msg=append(msg, "username must filled")
	} else if len(username) < 3 {
			msg=append(msg, "username must more than 3 characters")
	}


	if len(pass) == 0 {
		msg=append(msg, "password must filled")
	} else if len(pass) < 5 {
		msg=append(msg, "password must be more than 5")
	}
	

	if len(msg) > 0 {
		c.HTML(http.StatusOK,"register.html",gin.H{
			"msg" : msg,
		})
	} else {
		user.Email = email
		user.Username = username
		hash,_ := bcrypt.GenerateFromPassword([]byte(pass),16)
		user.Password = string(hash)
		db.Create(&user)
		// c.HTML(http.StatusOK,"login.html",gin.H{"":"",})
		c.Redirect(http.StatusMovedPermanently,"/login")
	}
}

func login(c *gin.Context){
	ses := sessions.Default(c)
	name := ses.Get("name")
	fmt.Println("masuk login")

	if name == nil{
		fmt.Println("blom")
		c.HTML(http.StatusOK,"login.html",gin.H{"":"",})
	} else {
		fmt.Println("udh")
		// item := GetItems(2)
		// c.HTML(http.StatusOK,"index.html",gin.H{"Item":item,"Name":name})
		c.Redirect(http.StatusMovedPermanently,"/")
	}
}

func register(c *gin.Context){
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		c.HTML(http.StatusOK,"register.html",gin.H{"":"",})
	} else {
		item := GetItems(2)
		c.HTML(http.StatusOK,"index.html",gin.H{"Item":item,"Name":name})
	}
}

func GetItems(n int) []Item{
	var item []Item
	if n < 5 {
		db.Where("categories_id = ?",n).Find(&item)
	} else {
		db.Find(&item)
	}

	for key:= range item{
			base := string(base64.StdEncoding.EncodeToString([]byte(item[key].ItemImage)))
			item[key].ItemImage = base
	}

	return item
}

func index(c *gin.Context){
	item := GetItems(2)
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}

	

	c.HTML(http.StatusOK,"index.html",gin.H{
		"Item":item,
		"Name" : name,
	})
}

func sayur(c *gin.Context){
	item := GetItems(1)
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}

	

	c.HTML(http.StatusOK,"sayur.html",gin.H{
		"Item":item,
		"Name" : name,
	})
}

func buah(c *gin.Context){
	item := GetItems(2)
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}

	

	c.HTML(http.StatusOK,"buah.html",gin.H{
		"Item":item,
		"Name" : name,
	})
}

func hidro(c *gin.Context){
	item := GetItems(3)
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}

	

	c.HTML(http.StatusOK,"hidro.html",gin.H{
		"Item":item,
		"Name" : name,
	})
}

func eksklu(c *gin.Context){
	item := GetItems(4)
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}

	

	c.HTML(http.StatusOK,"eksklu.html",gin.H{
		"Item":item,
		"Name" : name,
	})
}

func shop(c *gin.Context){
	item := GetItems(5)
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}

	

	c.HTML(http.StatusOK,"shop.html",gin.H{
		"Item":item,
		"Name" : name,
	})
}


func contact(c *gin.Context){
	ses := sessions.Default(c)
	name := ses.Get("name")

	if name == nil{
		name=""
	}
	c.HTML(http.StatusOK,"contact.html",gin.H{"Name":name,})
}