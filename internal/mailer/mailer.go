package mailer

const ActivationTemplate = "activation.html"

type Client interface {
	Send(templateFile string, username string, email string, data any) error
}
