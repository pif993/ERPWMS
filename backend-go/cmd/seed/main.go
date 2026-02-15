package main
import("context";"fmt";"os";"time";"github.com/google/uuid";"github.com/jackc/pgx/v5/pgxpool";"erpwms/backend-go/internal/common/crypto")
func getenv(k,d string)string{v:=os.Getenv(k);if v==""{return d};return v}
func must(err error){if err!=nil{panic(err)}}
func main(){
ctx:=context.Background()
dburl:=getenv("DB_URL","")
if dburl==""{panic("DB_URL missing")}
pool,err:=pgxpool.New(ctx,dburl);must(err);defer pool.Close()
email:=getenv("SEED_ADMIN_EMAIL","superadmin@erpwms.local")
pass:=getenv("SEED_ADMIN_PASSWORD","ChangeMe123!")
status:="active"
emailHash:=crypto.SearchHash(email,getenv("EMAIL_HASH_KEY","dev-email-hash-key-change"))
ph,err:=crypto.HashPassword(pass,crypto.DefaultArgon2Params());must(err)

var uid uuid.UUID
err=pool.QueryRow(ctx,`select id from users where email_hash=$1`,emailHash).Scan(&uid)
if err!=nil{
uid=uuid.New()
_,err=pool.Exec(ctx,`insert into users(id,email_hash,status,created_at) values($1,$2,$3,$4)`,uid,emailHash,status,time.Now());must(err)
}

_,err=pool.Exec(ctx,`insert into roles(id,name,created_at) values($1,'superadmin',$2) on conflict(name) do nothing`,uuid.New(),time.Now());must(err)
var rid uuid.UUID
must(pool.QueryRow(ctx,`select id from roles where name='superadmin'`).Scan(&rid))
_,err=pool.Exec(ctx,`insert into user_roles(user_id,role_id) values($1,$2) on conflict do nothing`,uid,rid);must(err)

_,err=pool.Exec(ctx,`insert into user_credentials(user_id,email,password_hash,created_at) values($1,$2,$3,$4)
on conflict(email) do update set user_id=excluded.user_id,password_hash=excluded.password_hash`,uid,email,ph,time.Now());must(err)

fmt.Printf("[seed] superadmin ready email=%s password=%s user_id=%s\n",email,pass,uid.String())
}
