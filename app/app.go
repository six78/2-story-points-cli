package app

//
//type App struct {
//	logger *zap.Logger
//	ctx    context.Context
//	cancel context.CancelFunc
//
//	waku                 *node.WakuNode
//	wakuConnectionStatus chan node.ConnStatus
//
//	session *Session
//}
//
//func NewApp() (*App, error) {
//	logger, err := zap.NewDevelopment()
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to configure logging")
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	waku, status, err := createWakuNode()
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to create waku node")
//	}
//
//	return &App{
//		logger:               logger,
//		ctx:                  ctx,
//		cancel:               cancel,
//		waku:                 waku,
//		wakuConnectionStatus: status,
//	}, nil
//}
//
//func (a *App) CreateNewSession(name string) error {
//	session, err := NewSession(true, name)
//	if err != nil {
//		return errors.Wrap(err, "failed to create new session")
//	}
//	a.session = session
//	// WARNING: start app here?
//}
//
//func (a *App) ConnectToSession(name string) error {
//	session, err := NewSession(false, name)
//	if err != nil {
//		return errors.Wrap(err, "failed to connect to session")
//	}
//	a.session = session
//	// WARNING: start app here?
//}
//
//func (a *App) Stop() {
//	a.cancel()
//	a.waku.Stop()
//}
