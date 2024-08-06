package main

import (
	"authorsdb-rest/db"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	docs "authorsdb-rest/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func CreateConnection() (*pgxpool.Conn, context.Context, *db.Queries, error) {
	ctx := context.Background()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	queries := db.New(conn)

	return conn, ctx, queries, nil
}

var pool *pgxpool.Pool

func main() {
	ctx := context.Background()
	tp, err := pgxpool.New(ctx, "postgres://postgres:Qweasdzxc123@localhost:5432/my-db?sslmode=disable")
	if err != nil {
		panic(err)
	}
	pool = tp

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	//r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	v1 := r.Group("/api/v1")
	{
		v1.GET("/authors", GetAuthors)
		v1.GET("/authors/:id", GetAuthor)
		v1.POST("/authors", CreateAuthor)
		v1.DELETE("/authors/:id", DeleteAuthor)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	http.ListenAndServe(":8080", r)
}

type AuthorModel struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateAuthorModel(a *db.Author) AuthorModel {
	return AuthorModel{a.ID, a.Name, a.Bio.String}
}

func CreateAuthorModelList(l []db.Author) []AuthorModel {
	res := make([]AuthorModel, len(l))
	for i, v := range l {
		res[i] = CreateAuthorModel(&v)
	}
	return res
}

// @Summary Get a list of authors
// @Description
// @Accept  json
// @Produce  json
// @Success 200 {array} AuthorModel "ok"
// @Router /authors [get]
func GetAuthors(c *gin.Context) {
	if conn, ctx, queries, err := CreateConnection(); err == nil {
		if authorList, e := queries.ListAuthors(ctx); e == nil {
			c.JSON(http.StatusOK, CreateAuthorModelList(authorList))
		} else {
			c.AbortWithError(http.StatusInternalServerError, e)
		}
		defer conn.Release()
	} else {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

// @Summary Get authors by Id
// @Description
// @Accept  json
// @Param id path int true "Id"
// @Produce  json
// @Success 200 {object} AuthorModel "ok"
// @Router /authors/{id} [get]
func GetAuthor(c *gin.Context) {
	idStr := c.Params.ByName("id")
	id, _ := strconv.Atoi(idStr)
	if conn, ctx, queries, err := CreateConnection(); err == nil {
		if author, e := queries.GetAuthor(ctx, int64(id)); e == nil {
			c.JSON(http.StatusOK, CreateAuthorModel(&author))
		} else {
			c.AbortWithError(http.StatusNotFound, e)
		}
		defer conn.Release()
	} else {
		c.AbortWithError(http.StatusUnprocessableEntity, err)
	}
}

func (r *AuthorCreateModel) BindAuthorModel(c *gin.Context, dbModel *db.CreateAuthorParams) error {
	if err := c.ShouldBindJSON(r); err != nil {
		return err
	}

	dbModel.Name = r.Name
	dbModel.Bio = pgtype.Text{String: r.Description, Valid: true}
	return nil
}

type AuthorCreateModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// @Summary Create authtor
// @Description
// @Accept  json
// @Param aut body AuthorCreateModel true "AuthorCreateModel"
// @Produce  json
// @Success 201 {object} AuthorModel "ok"
// @Success 422 {object} error "error"
// @Success 500 {object} error "internalError"
// @Router /authors [post]
func CreateAuthor(c *gin.Context) {
	var (
		req     AuthorCreateModel
		dbModel db.CreateAuthorParams
	)
	if er := req.BindAuthorModel(c, &dbModel); er == nil {
		if conn, ctx, queries, err := CreateConnection(); err == nil {
			if author, e := queries.CreateAuthor(ctx, dbModel); e == nil {
				c.JSON(http.StatusCreated, CreateAuthorModel(&author))
			} else {
				c.AbortWithError(http.StatusInternalServerError, e)
			}
			defer conn.Release()
		} else {
			c.AbortWithError(http.StatusUnprocessableEntity, err)
		}
	} else {
		c.AbortWithError(http.StatusUnprocessableEntity, er)
	}
}

// @Summary Delete author by Id
// @Description
// @Accept  json
// @Param id path int true "Id"
// @Produce  json
// @Success 200  "ok"
// @Success 404  "ok"
// @Router /authors/{id} [delete]
func DeleteAuthor(c *gin.Context) {
	idStr := c.Params.ByName("id")
	id, _ := strconv.Atoi(idStr)
	if conn, ctx, queries, err := CreateConnection(); err == nil {
		if v, e := queries.DeleteAuthor(ctx, int64(id)); e == nil && v > 0 {
			c.AbortWithStatus(http.StatusOK)
		} else {
			c.AbortWithError(http.StatusNotFound, e)
		}
		defer conn.Release()
	} else {
		c.AbortWithError(http.StatusNotFound, err)
	}
}
