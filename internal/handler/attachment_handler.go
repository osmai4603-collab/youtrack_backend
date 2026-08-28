package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"youtrack_backend/internal/domain"

	"github.com/google/uuid"
)

type AttachmentHandler struct {
	storageDir string
}

func NewAttachmentHandler(storageDir string) *AttachmentHandler {
	if storageDir == "" {
		storageDir = "./uploads"
	}
	_ = os.MkdirAll(storageDir, 0755)
	return &AttachmentHandler{
		storageDir: storageDir,
	}
}

// UploadAttachment يعالج رفع ملفات المرفقات للتذاكر POST /api/issues/{issueID}/attachments
func (h *AttachmentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	issueID := r.PathValue("issueID")
	if issueID == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			issueID = parts[3]
		}
	}

	// أقصى حجم للملف: 32MB
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse multipart form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "File field is required"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	fileID := uuid.New().String()
	fileName := fileID + ext
	filePath := filepath.Join(h.storageDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save file on server"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write file content"})
		return
	}

	attachment := &domain.Attachment{
		ID:        fileID,
		Name:      header.Filename,
		Size:      header.Size,
		MimeType:  header.Header.Get("Content-Type"),
		URL:       "/api/files/" + fileName,
		Created:   time.Now().UnixMilli(),
		Updated:   time.Now().UnixMilli(),
		Type:      "IssueAttachment",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(attachment)
}

// GetFile يعالج تحميل واستعراض الملفات GET /api/files/{fileName}
func (h *AttachmentHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("fileName")
	if fileName == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			fileName = parts[3]
		}
	}

	filePath := filepath.Join(h.storageDir, filepath.Clean(fileName))
	http.ServeFile(w, r, filePath)
}
