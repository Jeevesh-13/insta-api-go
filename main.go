package main

import (
	"context"
	"fmt"
	"log"
	"ioutil"
	"os"
	"net/http"
	"github.com/gorilla/mux"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	jwt "github.com/dgrijalva/jwt-go"
)


type user struct{
  ID          string `bson:"id" json:"ID,omitempty"`
  Name        string `bson:"Name" json:"Name"`
  Email       string `bson:"Email" json:"Email"`
  Password    string `bson:"Password" json:"Password"`
}
type allUsers []user
var users = allUsers{
  {
    ID:"1",
    Name:"Jeevesh",
    Email:"ghhs234@gmail.com",
    Password:"Prince@1387"
  },
}


type post struct{
  ID          string `bson:"id" json:"ID,omitempty"`
  Caption     string `bson:"Caption" json:"Caption"`
  URL         string `bson:"URL" json:"URL"`
  TIMESTAMP   time.Time `bson:"TIMESTAMP" json:"TIMESTAMP"`
}
type allPosts []post
var posts = allPosts{
  {
    ID:"1",
    Caption:"photo",
    URL:"https://www.pinterest.com/pin/729935052106223989/",
    TIMESTAMP:"09/10/2021 01:18:23 AM"
  },
}


func createUser(w http.ResponseWriter, r *http.Request, ctx context, database *gorm.DB){
  var newUser user
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newUser)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 8)
	newUser.Password := string(hashedPassword)
	users = append(users, newUser)
	w.WriteHeader(http.StatusCreated)
	usersCollection := database.Collection("users")
	result, err := usersCollection.InsertOne(ctx, newUser)
  if err != nil {
      fmt.Println(err)
      return
  }
	json.NewEncoder(w).Encode(newUser)
}

func getOneUser(w http.ResponseWriter, r *http.Request, ctx context, database *gorm.DB) {
	usersCollection := database.Collection("users")
	findResult := usersCollection.FindOne(ctx, bson.M{"id": mux.Vars(r)["id"]})
  if err := findResult.Err(); err != nil {
      fmt.Println(err)
      return
  }
  n := user{}
  err = findResult.Decode(&n)
  if err != nil {
      fmt.Println(err)
      return
  }
  fmt.Println(n.Body)
	json.NewEncoder(w).Encode(n)
}

func getOneIdPost(w http.ResponseWriter, r *http.Request, ctx context, database *gorm.DB) {
	postsCollection := database.Collection("posts")
	findResult := postsCollection.FindOne(ctx, bson.M{"id": mux.Vars(r)["id"]})
  if err := findResult.Err(); err != nil {
      fmt.Println(err)
      return
  }
  n := post{}
  err = findResult.Decode(&n)
  if err != nil {
      fmt.Println(err)
      return
  }
  fmt.Println(n.Body)
	json.NewEncoder(w).Encode(n)
}

func getAllPosts(w http.ResponseWriter, r *http.Request, ctx context, database *gorm.DB) {
	postsResult := allPosts{}
	postsCollection := database.Collection("posts")
  cursor, err := postsCollection.Find(ctx, bson.M{"id": mux.Vars(r)["id"]})
  if err != nil {
      fmt.Println(err)
      return
  }
  for cursor.Next(ctx) {
		n := post{}
		cursor.Decode(&n)
		postsResult = append(postsResult, n)
  }
  for _, el := range postsResult {
		fmt.Println(el.Body)
  }
	json.NewEncoder(w).Encode(postsResult)
}

func createPost(w http.ResponseWriter, r *http.Request, ctx context, database *gorm.DB){
  var newPost post
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newPost)
	events = append(posts, newPost)
	w.WriteHeader(http.StatusCreated)

	postsCollection := database.Collection("posts")
	result, err := coll.InsertOne(ctx, newPost)
  if err != nil {
      fmt.Println(err)
      return
  }
	json.NewEncoder(w).Encode(newPost)
}

var mySigningKey = []byte("captain")

//for checking server api is secure
func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
        }
        return mySigningKey, nil
    	})
			if err != nil {
				fmt.Fprintf(w, err.Error())
      }
			if token.Valid {
				endpoint(w, r)
			}
    } else {
			fmt.Fprintf(w, "Not Authorized")
      }
		})
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	validToken, err := GenerateJWT()
	if err != nil {
		fmt.Println("Failed to generate token")
	}
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:9000/", nil)
	req.Header.Set("Token", validToken)
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(body))
}

//to make api secure
func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["client"] = "Elliot Forbes"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
	  return "", err
	}
	return tokenString, nil
}

func main() {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("ATLAS_URI")))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	database := client.Database("quickinsta")
	postsCollection := database.Collection("posts")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
  router.HandleFunc("/users",func(w http.ResponseWriter, r *http.Request){
		createUser(w, r, ctx, database)
		}).Method("POST")
  router.HandleFunc("/users/{id}",func (w http.ResponseWriter, r *http.Request)  {
  	getOneUser(w, r, ctx, database)
  }).Method("GET")
  router.HandleFunc("/posts",func (w http.ResponseWriter, r *http.Request){
  	createPost(w, r, ctx, database)
		}).Method("POST")
  router.HandleFunc("/posts/{id}",func (w http.ResponseWriter, r *http.Request)  {
  	getOneIdPost(w, r, ctx, database)
  }).Method("GET")
  router.HandleFunc("/posts/users/{id}",func (w http.ResponseWriter, r *http.Request)  {
  	getAllPosts(w, r, ctx, database)
		}).Method("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
