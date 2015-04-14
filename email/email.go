package email

import (
  "net/smtp"
  "net/mail"
  "crypto/tls"
  "errors"
  "log"
  "fmt"
)

type Logger interface {
  Trace(format string, v ...interface{})
  Debug(format string, v ...interface{})
  Info(format string, v ...interface{})
  Warn(format string, v ...interface{})
  Error(format string, v ...interface{})
  Fatal(format string, v ...interface{})
}

type EmailConfig struct {
  Host string
  Port int16
  SSL bool
  User string
  Password string
  Sender string
}

func NewClientMailer(config EmailConfig) ClientEmailer {
  return ClientEmailer{config}
}

type ClientEmailer struct {
  config EmailConfig
}

func (ce ClientEmailer) sendSSLEmail( address string, content string) error {
  from  := mail.Address{"", ce.config.Sender}
  to    := mail.Address{"", address}
  subj  := "MGM Email"
  body  := content

  headers := make(map[string]string)
  headers["From"] = from.String()
  headers["To"] = to.String()
  headers["Subject"] = subj

  message := ""
  for k,v := range headers {
    message += fmt.Sprintf("%s: %s\r\n", k, v)
  }
  message += "\r\n" + body

  servername := fmt.Sprintf("%v:%v", ce.config.Host, ce.config.Port)

  auth := smtp.PlainAuth("", ce.config.User, ce.config.Password, ce.config.Host)

  tlsconfig := &tls.Config {
    InsecureSkipVerify: true,
    ServerName: ce.config.Host,
  }

  conn, err := tls.Dial("tcp", servername, tlsconfig)
  if err != nil {
    log.Panic(err)
  }

  c, err := smtp.NewClient(conn, ce.config.Host)
  if err != nil {
    log.Panic(err)
  }

  if err = c.Auth(auth); err != nil {
    log.Panic(err)
  }

  if err = c.Mail(from.Address); err != nil {
    log.Panic(err)
  }

  if err = c.Rcpt(to.Address); err != nil {
    log.Panic(err)
  }

  w, err := c.Data()
  if err != nil {
    log.Panic(err)
  }

  _, err = w.Write([]byte(message))
  if err != nil {
    log.Panic(err)
  }

  err = w.Close()
  if err != nil {
    log.Panic(err)
  }

  c.Quit()
  return nil
}

func (ce ClientEmailer) sendNonSSLEmail( address string, message string) error {
  return errors.New("Non SSL email is not implemented")
}

func (ce ClientEmailer) TestMessage (address string, message string) error{

  if ce.config.SSL {
    return ce.sendSSLEmail(address, message)
  } else {
    return ce.sendNonSSLEmail(address, message)
  }
}
