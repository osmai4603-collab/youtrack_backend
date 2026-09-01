package app

type Option func(s *YouTrackServer) error

type (
	AppOption        func(a *YouTrackApp)
	AppOptionCreator func() []AppOption
)
