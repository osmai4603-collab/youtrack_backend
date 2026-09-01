package model

// UserConfig يمثّل الإعدادات المستخدمة لتشغيل الخادم (من env).
// هذا هو المصدر المطابق لملف .env.example.
type ServerConfig struct {
	ServerPort string
	AppEnv     string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
}

// DSN يبني سلسلة الاتصال بقاعدة البيانات.
func (c *ServerConfig) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword +
		"@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName +
		"?sslmode=" + c.DBSSLMode
}

// Port يعيد منفذ الاستماع بصيغة ":"+port.
func (c *ServerConfig) Port() string {
	return ":" + c.ServerPort
}
