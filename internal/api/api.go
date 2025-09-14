package api

type App struct {
	db interface{}
}

func New(db interface{}) *App {
	return &App{db: db}
}

func (a *App) Listen(addr string) error {
	return nil
}
