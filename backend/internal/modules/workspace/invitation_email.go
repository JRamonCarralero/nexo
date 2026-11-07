package workspace

import (
	"bytes"
	htmltemplate "html/template"
	texttemplate "text/template"
	"time"

	"github.com/JRamonCarralero/nexo/backend/internal/platform/email"
)

// The email is bilingual (Spanish, then English): the invitee's language is
// not known yet.

type invitationEmailData struct {
	To        string
	Inviter   string
	Workspace string
	Role      Role
	URL       string
	ExpiresAt time.Time
}

func (d invitationEmailData) RoleES() string {
	switch d.Role {
	case RoleAdmin:
		return "administrador/a"
	case RoleGuest:
		return "invitado/a"
	default:
		return "miembro"
	}
}

func (d invitationEmailData) RoleEN() string {
	switch d.Role {
	case RoleAdmin:
		return "admin"
	case RoleGuest:
		return "guest"
	default:
		return "member"
	}
}

func (d invitationEmailData) Expires() string { return d.ExpiresAt.UTC().Format("2006-01-02") }

var invitationSubject = texttemplate.Must(texttemplate.New("subject").Parse(
	`{{.Inviter}} te ha invitado a {{.Workspace}} en Nexo · {{.Inviter}} invited you to {{.Workspace}} on Nexo`))

var invitationText = texttemplate.Must(texttemplate.New("text").Parse(`Hola:

{{.Inviter}} te ha invitado a unirte a «{{.Workspace}}» en Nexo como {{.RoleES}}.

Acepta la invitación aquí (caduca el {{.Expires}}):
{{.URL}}

Si no esperabas esta invitación, puedes ignorar este correo.

---

Hi,

{{.Inviter}} has invited you to join "{{.Workspace}}" on Nexo as {{.RoleEN}}.

Accept the invitation here (it expires on {{.Expires}}):
{{.URL}}

If you were not expecting this invitation, you can ignore this email.
`))

var invitationHTML = htmltemplate.Must(htmltemplate.New("html").Parse(`<!doctype html>
<html><body style="font-family: system-ui, sans-serif; color: #0f172a; line-height: 1.5">
<p>Hola:</p>
<p><strong>{{.Inviter}}</strong> te ha invitado a unirte a <strong>{{.Workspace}}</strong> en Nexo como {{.RoleES}}.</p>
<p><a href="{{.URL}}" style="display: inline-block; padding: 10px 16px; background: #0f172a; color: #ffffff; border-radius: 6px; text-decoration: none">Aceptar la invitación</a></p>
<p style="color: #475569; font-size: 14px">Caduca el {{.Expires}}. Si no esperabas esta invitación, puedes ignorar este correo.</p>
<hr style="border: none; border-top: 1px solid #e2e8f0; margin: 24px 0">
<p>Hi,</p>
<p><strong>{{.Inviter}}</strong> has invited you to join <strong>{{.Workspace}}</strong> on Nexo as {{.RoleEN}}.</p>
<p><a href="{{.URL}}" style="display: inline-block; padding: 10px 16px; background: #0f172a; color: #ffffff; border-radius: 6px; text-decoration: none">Accept the invitation</a></p>
<p style="color: #475569; font-size: 14px">It expires on {{.Expires}}. If you were not expecting this invitation, you can ignore this email.</p>
</body></html>
`))

func invitationEmail(d invitationEmailData) (email.Message, error) {
	var subject, text, html bytes.Buffer
	if err := invitationSubject.Execute(&subject, d); err != nil {
		return email.Message{}, err
	}
	if err := invitationText.Execute(&text, d); err != nil {
		return email.Message{}, err
	}
	if err := invitationHTML.Execute(&html, d); err != nil {
		return email.Message{}, err
	}
	return email.Message{To: d.To, Subject: subject.String(), Text: text.String(), HTML: html.String()}, nil
}
