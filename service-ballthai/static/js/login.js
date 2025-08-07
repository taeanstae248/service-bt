// Login Page JavaScript
const API_BASE_URL = window.location.protocol + '//' + window.location.host;

document.addEventListener('DOMContentLoaded', function() {
    // Check if already logged in
    const sessionId = localStorage.getItem('sessionId');
    if (sessionId) {
        // Verify session is still valid
        fetch(`${API_BASE_URL}/api/auth/verify`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${sessionId}`
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                window.location.href = '/dashboard.html';
            } else {
                localStorage.removeItem('sessionId');
                localStorage.removeItem('user');
            }
        })
        .catch(error => {
            console.error('Session verification failed:', error);
            localStorage.removeItem('sessionId');
            localStorage.removeItem('user');
        });
    }

    // Handle form submission
    const loginForm = document.getElementById('loginForm');
    loginForm.addEventListener('submit', handleLogin);
});

async function handleLogin(event) {
    event.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const submitBtn = document.getElementById('submitBtn');
    const loadingDiv = document.getElementById('loading');
    
    // Validation
    if (!username || !password) {
        showError('กรุณากรอกชื่อผู้ใช้และรหัสผ่าน');
        return;
    }
    
    // Show loading state
    submitBtn.disabled = true;
    submitBtn.textContent = 'กำลังเข้าสู่ระบบ...';
    loadingDiv.style.display = 'block';
    hideMessages();
    
    try {
        console.log('Attempting login with:', username);
        const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        });
        
        console.log('Response status:', response.status);
        const data = await response.json();
        console.log('Response data:', data);
        
        if (response.ok && data.success) {
            // Store session info
            localStorage.setItem('sessionId', data.session_id);
            localStorage.setItem('user', JSON.stringify(data.user));
            
            showSuccess('เข้าสู่ระบบสำเร็จ! กำลังเปลี่ยนหน้า...');
            
            setTimeout(() => {
                window.location.href = '/dashboard.html';
            }, 1000);
            
        } else {
            showError(data.error || 'การเข้าสู่ระบบล้มเหลว กรุณาลองใหม่อีกครั้ง');
        }
        
    } catch (error) {
        console.error('Login error:', error);
        showError('เกิดข้อผิดพลาดในการเชื่อมต่อ กรุณาลองใหม่อีกครั้ง');
    } finally {
        // Reset button state
        submitBtn.disabled = false;
        submitBtn.textContent = 'เข้าสู่ระบบ';
        loadingDiv.style.display = 'none';
    }
}

function showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

function showSuccess(message) {
    const successDiv = document.getElementById('successMessage');
    successDiv.textContent = message;
    successDiv.style.display = 'block';
}

function hideMessages() {
    document.getElementById('errorMessage').style.display = 'none';
    document.getElementById('successMessage').style.display = 'none';
}
