package controllers

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dickson7/proyectogocomentarios/commons"
	"github.com/dickson7/proyectogocomentarios/configuration"
	"github.com/dickson7/proyectogocomentarios/models"
)

// Login es el controlador de login
func Login(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	// con esta linea se mapean los datos de r en user
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		fmt.Fprintf(w, "Error: %s\n", err)
		return
	}
	db := configuration.GetConnection()
	defer db.Close()

	c := sha256.Sum256([]byte(user.Password))
	//pwd := base64.URLEncoding.EncodeToString(c[:32])
	pwd := fmt.Sprintf("%x", c)

	//todos los dats que lleguen del resultado de la consulta se mapean en user
	db.Where("email = ? and password = ?", user.Email, pwd).First(&user)
	if user.ID > 0 {
		// le damos valor vacio a password para seguridad y generamos el token
		user.Password = ""
		token := commons.GenerateJWT(user)

		j, err := json.Marshal(models.Token{Token: token})
		if err != nil {
			log.Fatalf("Error al convertir el token a json: %s", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	} else {
		m := models.Message{
			Message: "Usuario o clave no valido",
			Code:    http.StatusUnauthorized,
		}
		commons.DisplayMessage(w, m)
	}

}

// UserCreate permite registrar un usuario
func UserCreate(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	m := models.Message{}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		m.Message = fmt.Sprintf("Error al leer usuario a registrar: %s", err)
		m.Code = http.StatusBadRequest
		commons.DisplayMessage(w, m)
		return
	}

	if user.Password != user.ConfirmPassword {
		m.Message = "LLas contraseñas no coiciden"
		m.Code = http.StatusBadRequest
		commons.DisplayMessage(w, m)
		return
	}

	c := sha256.Sum256([]byte(user.Password))
	pwd := fmt.Sprintf("%x", c)
	user.Password = pwd
	picmd5 := md5.Sum([]byte(user.Email))
	picstr := fmt.Sprintf("%x", picmd5)
	user.Picture = "https://gravatar.com/avatar/" + picstr + "?s=100"

	db := configuration.GetConnection()
	defer db.Close()

	err = db.Create(&user).Error
	if err != nil {
		m.Message = fmt.Sprintf("Error al crear el registro %s", err)
		m.Code = http.StatusBadRequest
		commons.DisplayMessage(w, m)
		return
	}
	m.Message = "Usuario creado con exito"
	m.Code = http.StatusCreated
	commons.DisplayMessage(w, m)
}
