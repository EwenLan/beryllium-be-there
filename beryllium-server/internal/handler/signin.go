package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/crypto"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/model"
	"github.com/EwenLan/beryllium-be-there/beryllium-server/internal/store"
)

// SigninHandler handles the student sign-in flow.
type SigninHandler struct {
	classStore      *store.ClassStore
	studentStore    *store.StudentStore
	attendanceStore *store.AttendanceStore
}

// NewSigninHandler creates a SigninHandler.
func NewSigninHandler(classStore *store.ClassStore, studentStore *store.StudentStore, attendanceStore *store.AttendanceStore) *SigninHandler {
	return &SigninHandler{classStore: classStore, studentStore: studentStore, attendanceStore: attendanceStore}
}

// Page handles GET /signin/{class-id} — serves the sign-in page with injected public key.
func (h *SigninHandler) Page(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("classid")

	class, err := h.classStore.GetByID(classID)
	if err != nil {
		http.Error(w, "class not found", http.StatusNotFound)
		return
	}

	// Try to serve the built signin page; fall back to a simple inline page
	htmlPath := "../beryllium-signin/out/index.html"
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		// Fallback inline sign-in page for development
		html = []byte(fallbackSigninHTML)
	}

	// Inject class ID and public key before </head>
	script := fmt.Sprintf(`<script>window.__CLASS_ID__="%s";window.__CLASS_PUBLIC_KEY__="%s";</script>`, class.ClassID, class.PublicKey)
	injected := strings.Replace(string(html), "</head>", script+"</head>", 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(injected))
}

// Submit handles POST /signin/{class-id} — processes a sign-in submission.
func (h *SigninHandler) Submit(w http.ResponseWriter, r *http.Request) {
	classID := r.PathValue("classid")

	class, err := h.classStore.GetByID(classID)
	if err != nil {
		http.Error(w, `{"error":"class not found"}`, http.StatusNotFound)
		return
	}

	var req model.SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Decrypt password using class private key
	privKey, err := crypto.ParsePrivateKey(class.PrivateKey)
	if err != nil {
		http.Error(w, `{"error":"failed to parse private key"}`, http.StatusInternalServerError)
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		http.Error(w, `{"error":"invalid encrypted password encoding"}`, http.StatusBadRequest)
		return
	}

	plainPassword, err := crypto.DecryptPKCS1v15(privKey, ciphertext)
	if err != nil {
		http.Error(w, `{"error":"failed to decrypt password"}`, http.StatusBadRequest)
		return
	}

	// Look up student and verify password
	student, err := h.studentStore.GetByAccount(req.Account)
	if err != nil {
		http.Error(w, `{"error":"invalid account or password"}`, http.StatusUnauthorized)
		return
	}

	if !crypto.VerifyPassword(student.PasswordHash, string(plainPassword)) {
		http.Error(w, `{"error":"invalid account or password"}`, http.StatusUnauthorized)
		return
	}

	// Mark as present
	if err := h.attendanceStore.MarkPresent(classID, student.Account); err != nil {
		// If already signed in, still return success
		if err.Error() != "already signed in" {
			http.Error(w, `{"error":"failed to record attendance"}`, http.StatusInternalServerError)
			return
		}
	}

	type signinResponse struct {
		Success bool   `json:"success"`
		Name    string `json:"name"`
	}
	writeJSON(w, http.StatusOK, signinResponse{Success: true, Name: student.Name})
}

// fallbackSigninHTML is a basic sign-in page served when the built frontend is not available.
const fallbackSigninHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>课堂签到</title>
<script src="https://cdn.jsdelivr.net/npm/jsencrypt@3.3.2/bin/jsencrypt.min.js"></script>
<style>
body{font-family:system-ui,sans-serif;max-width:400px;margin:60px auto;padding:20px}
h1{text-align:center;color:#333}
form{display:flex;flex-direction:column;gap:12px}
input{padding:10px;font-size:16px;border:1px solid #ccc;border-radius:6px}
button{padding:12px;font-size:16px;background:#007bff;color:#fff;border:none;border-radius:6px;cursor:pointer}
button:hover{background:#0056b3}
#msg{margin-top:16px;text-align:center;font-weight:bold}
.success{color:#28a745}.error{color:#dc3545}
</style>
</head>
<body>
<h1>课堂签到</h1>
<p style="text-align:center;color:#666">Class: <span id="classId"></span></p>
<form id="signinForm">
<input type="text" id="account" placeholder="请输入账号" required>
<input type="password" id="password" placeholder="请输入密码" required>
<button type="submit">签到</button>
</form>
<div id="msg"></div>
<script>
document.getElementById("classId").textContent=window.__CLASS_ID__||"";
function encryptPassword(pubKeyPEM,password){var enc=new JSEncrypt();enc.setPublicKey(pubKeyPEM);var ct=enc.encrypt(password);if(!ct)throw new Error("encryption failed");return ct}
document.getElementById("signinForm").addEventListener("submit",async(e)=>{e.preventDefault();var msg=document.getElementById("msg");msg.className="";msg.textContent="签到中...";try{var encPwd=encryptPassword(window.__CLASS_PUBLIC_KEY__,document.getElementById("password").value);var res=await fetch(window.location.pathname,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({account:document.getElementById("account").value,password:encPwd})});var data=await res.json();if(res.ok){msg.className="success";msg.textContent="签到成功！"+data.name}else{msg.className="error";msg.textContent=data.error||"签到失败"}}catch(err){msg.className="error";msg.textContent="签到失败: "+err.message}})
</script>
</body>
</html>`
