package http
import("net/http";"github.com/gin-gonic/gin";"github.com/jackc/pgx/v5/pgxpool")
type AdminPortal struct{DB *pgxpool.Pool}
type U struct{ID string `json:"id"`;EmailHash string `json:"email_hash"`;Status string `json:"status"`}
type R struct{ID string `json:"id"`;Name string `json:"name"`}
func (p AdminPortal) RegisterRoutes(r *gin.Engine){
g:=r.Group("/admin")
g.GET("/users",p.Users)
g.GET("/roles",p.Roles)
g.POST("/users/:id/roles/:role_id",p.SetRole)
}
func (p AdminPortal) Users(c *gin.Context){
rows,err:=p.DB.Query(c,`select id::text,email_hash,status from users order by created_at desc limit 500`)
if err!=nil{c.JSON(500,gin.H{"error":err.Error()});return}
defer rows.Close()
out:=[]U{}
for rows.Next(){var x U;if e:=rows.Scan(&x.ID,&x.EmailHash,&x.Status);e!=nil{c.JSON(500,gin.H{"error":e.Error()});return};out=append(out,x)}
c.JSON(200,out)
}
func (p AdminPortal) Roles(c *gin.Context){
rows,err:=p.DB.Query(c,`select id::text,name from roles order by name`)
if err!=nil{c.JSON(500,gin.H{"error":err.Error()});return}
defer rows.Close()
out:=[]R{}
for rows.Next(){var x R;if e:=rows.Scan(&x.ID,&x.Name);e!=nil{c.JSON(500,gin.H{"error":e.Error()});return};out=append(out,x)}
c.JSON(200,out)
}
func (p AdminPortal) SetRole(c *gin.Context){
uid:=c.Param("id");rid:=c.Param("role_id")
tx,err:=p.DB.Begin(c);if err!=nil{c.JSON(500,gin.H{"error":err.Error()});return}
defer tx.Rollback(c)
if _,err=tx.Exec(c,`delete from user_roles where user_id=$1::uuid`,uid);err!=nil{c.JSON(500,gin.H{"error":err.Error()});return}
if _,err=tx.Exec(c,`insert into user_roles(user_id,role_id) values($1::uuid,$2::uuid) on conflict do nothing`,uid,rid);err!=nil{c.JSON(500,gin.H{"error":err.Error()});return}
if err=tx.Commit(c);err!=nil{c.JSON(500,gin.H{"error":err.Error()});return}
c.JSON(http.StatusOK,gin.H{"ok":true})
}
