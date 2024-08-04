package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"social-network-server/pkg/database"
	"social-network-server/pkg/models"
	"strconv"

	"social-network-server/pkg/models/errs"

	"github.com/gin-gonic/gin"
)

// Feed retrieves posts for the current user's feed.
func Feed(c *gin.Context) {
	// Obter o ID do usuário da sessão JWT
	id, exists := c.Get("id")
	if !exists {
		// Lidar com o caso em que o ID do usuário não está disponível
		log.Println("ID do usuário não encontrado na sessão")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário não encontrado na sessão"})
		return
	}

	db := database.GetDB()

	var post models.UserPost
	var chatPartnerUsername string
	var chatPartnerIcon []byte

	posts := []models.UserPost{}

	query := `
    SELECT user_post.postID, user_post.id AS post_user_id, user_post.content,
           user.id AS user_id, user.username, user.name, user.icon
    FROM user_post
    JOIN user ON user.id = user_post.id
    WHERE user.id = ? OR user.id IN (
        SELECT user_follow.followTo
        FROM user_follow
        WHERE user_follow.followBy = ?
    )
    ORDER BY user_post.created_at DESC
`

	rows, err := db.Query(query, id, id)
	if err != nil {
		log.Println("Failed to query statement", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute query",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var icon []byte

		err := rows.Scan(&post.PostID, &post.PostUserID, &post.Content, &post.UserID, &post.CreatedBy, &post.Name, &icon)
		if err != nil {
			log.Println("Failed to scan statement", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan rows",
			})
			return
		}

		var imageBase64 string
		if icon != nil {
			imageBase64 = base64.StdEncoding.EncodeToString(icon)
		}

		posts = append(posts, models.UserPost{
			PostID:     post.PostID,
			PostUserID: post.PostUserID,
			Content:    post.Content,
			UserID:     post.UserID,
			CreatedBy:  post.CreatedBy,
			Name:       post.Name,
			IconBase64: imageBase64,
		})
	}
	log.Println("Number of posts retrieved:", len(posts)) // Log para verificar o número de posts

	// Obter o user e o ícone do usuário
	err = db.QueryRow("SELECT username, icon FROM user WHERE id = ?", id).Scan(&chatPartnerUsername, &chatPartnerIcon)
	if err != nil {
		log.Println("Failed to query chat partner details:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get chat partner details"})
		return
	}
	var chatPartnerIconBase64 string
	if chatPartnerIcon != nil {
		chatPartnerIconBase64 = base64.StdEncoding.EncodeToString(chatPartnerIcon)
	}

	post.IconBase64 = chatPartnerIconBase64

	c.JSON(http.StatusOK, gin.H{
		"posts":       posts,
		"chatPartner": gin.H{"username": chatPartnerUsername, "iconBase64": chatPartnerIconBase64},
	})
}

func CreateNewPost(c *gin.Context) {
	var userPost models.UserPost
	errresp := errs.ErrorResponse{
		Error: make(map[string]string),
	}

	// Obter o ID do usuário da sessão JWT
	userID, exists := c.Get("id")
	if !exists {
		log.Println("ID do usuário não encontrado na sessão")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário não encontrado na sessão"})
		return
	}

	// Ler o corpo da requisição JSON
	if err := c.ShouldBindJSON(&userPost); err != nil {
		log.Println("Erro ao vincular JSON:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados de post inválidos"})
		return
	}

	log.Printf("Recebido conteúdo do post: %s", userPost.Content)

	// Validar conteúdo do post
	if userPost.Content == "" {
		errresp.Error["content"] = "O conteúdo do post não pode estar vazio!"
	}

	if len(errresp.Error) > 0 {
		c.JSON(http.StatusBadRequest, errresp)
		return
	}

	id, errId := strconv.Atoi(fmt.Sprintf("%v", userID))
	if errId != nil {
		log.Println("Erro ao converter ID do usuário para int:", errId)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário inválido"})
		c.Abort()
		return
	}

	// Obter nome de usuário
	var username string
	db := database.GetDB()
	err := db.QueryRow("SELECT username FROM user WHERE id = ?", id).Scan(&username)
	if err != nil {
		log.Println("Error querying username:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha ao consultar nome de usuário",
		})
		return
	}

	// Preencher dados do post
	userPost.CreatedBy = username

	// Executar inserção no banco de dados
	stmt, err := db.Prepare("INSERT INTO user_post(content, createdBy, id, created_at) VALUES (?, ?, ?, NOW())")
	if err != nil {
		log.Println("Erro ao preparar declaração SQL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha ao preparar declaração",
		})
		return
	}

	rs, err := stmt.Exec(userPost.Content, userPost.CreatedBy, id)
	if err != nil {
		log.Println("Erro ao executar declaração SQL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Falha ao executar declaração",
		})
		return
	}

	insertID, _ := rs.LastInsertId()

	resp := map[string]interface{}{
		"postID":  insertID,
		"message": "Post criado com sucesso!",
	}
	c.JSON(http.StatusOK, resp)
}

// DeletePost deletes a post based on its ID.
func DeletePost(c *gin.Context) {
	postID := c.PostForm("post")
	// Obter o ID do usuário da sessão JWT
	id, exists := c.Get("id")
	if !exists {
		// Lidar com o caso em que o ID do usuário não está disponível
		log.Println("ID do usuário não encontrado na sessão")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID do usuário não encontrado na sessão"})
		return
	}

	db := database.GetDB()

	var postAuthorID int
	err := db.QueryRow("SELECT id FROM user_post WHERE postID=?", postID).Scan(&postAuthorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch post details",
		})
		return
	}

	if postAuthorID != id {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have permission to delete this post",
		})
		return
	}

	_, err = db.Exec("DELETE FROM user_post WHERE postID=?", postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete post",
		})
		return
	}

	resp := map[string]interface{}{
		"mssg": "Post Deleted!",
	}
	c.JSON(http.StatusOK, resp)
}
