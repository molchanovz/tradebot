package app

//
//import (
//	"github.com/labstack/echo/v4"
//	"github.com/vmkteam/vfs"
//	vfsdb "github.com/vmkteam/vfs/db"
//	zm "github.com/vmkteam/zenrpc-middleware"
//	"net/http"
//	"tradebot/pkg/db"
//)
//
//const NSVFS = "vfs"
//
//// RegisterVFS register VFS handler and RPC service
//func (a *App) RegisterVFS(cfg vfs.Config) error {
//	vf, err := vfs.New(cfg)
//	if err != nil {
//		return err
//	}
//
//	cr := db.NewTradebotRepo(a.db)
//	vfsRepo := vfsdb.NewVfsRepo(a.db)
//	a.echo.Any("/v1/vfs/upload/file", zm.EchoHandler(vt.HTTPAuthMiddleware(cr, vf.UploadHandler(vfsRepo))))
//	a.echo.Any("/v1/vfs/upload/hash", echo.WrapHandler(vt.HTTPAuthMiddleware(cr, vf.HashUploadHandler(&vfsRepo))))
//	a.echo.get(a.cfg.VFS.WebPath, echo.WrapHandler(http.StripPrefix(a.cfg.VFS.WebPath, http.FileServer(http.Dir(a.cfg.VFS.Path)))))
//	vt.WebPath = a.cfg.VFS.WebPath
//
//	a.vtsrv.Register(NSVFS, vfs.NewService(vfsRepo, vf, a.dbc))
//
//	return nil
//}
