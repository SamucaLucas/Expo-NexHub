package services

import (
	"fmt"
	"net/smtp"
	"os"
)

func EnviarEmailRecuperacao(destinatario, codigo string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	// Pega a URL do site.
	// IMPORTANTE: Para a imagem aparecer no email, essa URL precisa ser PÚBLICA (ex: https://meusite.com)
	// Se for localhost, a imagem vai quebrar no Gmail/Outlook.
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("credenciais de e-mail não configuradas")
	}

	// Headers essenciais para evitar cair em SPAM e aceitar HTML
	assunto := "Subject: Redefinicao de Senha - NexHub\n"
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	msgHeader := []byte(assunto + headers + "\n")

	// HTML RESPONSIVO E MODERNO
	// %s (1º) = baseURL (Link da imagem)
	// %s (2º) = codigo (O número)
	corpoHTML := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
    </head>
    <body style="margin: 0; padding: 0; background-color: #f4f4f5; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;">
        
        <table role="presentation" border="0" cellpadding="0" cellspacing="0" width="100%%">
            <tr>
                <td style="padding: 20px 10px;">
                    
                    <div style="max-width: 480px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 10px rgba(0,0,0,0.05);">
                        
                        <div style="background-color: #18181b; padding: 30px 20px; text-align: center;">
                            <img src="%s/static/img/logo1.png" alt="NexHub" style="height: auto; width: 120px; max-width: 100%%; display: block; margin: 0 auto;">
                        </div>

                        <div style="padding: 30px 25px; text-align: center; color: #333;">
                            
                            <h2 style="color: #18181b; margin-top: 0; font-size: 22px;">Recuperação de Senha</h2>
                            <p style="color: #52525b; font-size: 16px; line-height: 1.5; margin-bottom: 25px;">
                                Recebemos uma solicitação para redefinir a senha da sua conta no <strong>NexHub</strong>.
                            </p>
                            
                            <div style="background: #ecfdf5; border: 1px dashed #10b981; color: #047857; padding: 20px; border-radius: 8px; margin: 30px 0;">
                                <span style="display: block; font-size: 14px; text-transform: uppercase; font-weight: 600; margin-bottom: 10px; color: #059669;">Seu Código de Acesso</span>
                                <strong style="font-size: 36px; letter-spacing: 6px; font-family: monospace; display: block;">%s</strong>
                            </div>

                            <p style="color: #71717a; font-size: 14px; margin-top: 25px;">Este código é válido por <strong>15 minutos</strong>.</p>
                            
                            <div style="height: 1px; background-color: #e4e4e7; margin: 30px 0;"></div>

                            <p style="font-size: 13px; color: #a1a1aa; margin: 0;">
                                Se você não solicitou essa alteração, por favor ignore este e-mail. Sua conta permanece segura.
                            </p>
                        </div>

                        <div style="background-color: #fafafa; padding: 15px; text-align: center; font-size: 12px; color: #a1a1aa; border-top: 1px solid #f4f4f5;">
                            &copy; 2026 NexHub. Todos os direitos reservados.
                        </div>

                    </div>
                </td>
            </tr>
        </table>

    </body>
    </html>
    `, baseURL, codigo)

	msg := append(msgHeader, []byte(corpoHTML)...)

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{destinatario}, msg)

	return err
}
