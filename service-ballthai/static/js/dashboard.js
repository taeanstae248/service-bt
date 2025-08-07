// Dashboard Page JavaScript
const API_BASE_URL = window.location.protocol + '//' + window.location.host;

// Check authentication on page load
window.addEventListener('load', function() {
    checkAuth();
    loadDashboardData();
});

function checkAuth() {
    const sessionId = localStorage.getItem('sessionId');
    const user = JSON.parse(localStorage.getItem('user') || '{}');

    if (!sessionId) {
        window.location.href = '/login.html';
        return;
    }

    // Display user info
    const userWelcome = document.getElementById('userWelcome');
    if (user.full_name) {
        userWelcome.textContent = `ยินดีต้อนรับ, ${user.full_name}`;
    } else if (user.username) {
        userWelcome.textContent = `ยินดีต้อนรับ, ${user.username}`;
    }

    // Verify session is still valid
    fetch(`${API_BASE_URL}/api/auth/verify`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${sessionId}`
        }
    })
    .then(response => response.json())
    .then(data => {
        if (!data.success) {
            logout();
        }
    })
    .catch(error => {
        console.error('Auth verification failed:', error);
        logout();
    });
}

async function loadDashboardData() {
    try {
        // Load statistics
        const [leagues, teams, players, matches] = await Promise.all([
            fetch(`${API_BASE_URL}/api/leagues`).then(r => r.json()),
            fetch(`${API_BASE_URL}/api/teams`).then(r => r.json()),
            fetch(`${API_BASE_URL}/api/players?limit=1000`).then(r => r.json()),
            fetch(`${API_BASE_URL}/api/matches?limit=1000`).then(r => r.json())
        ]);

        document.getElementById('totalLeagues').textContent = leagues.data ? leagues.data.length : 0;
        document.getElementById('totalTeams').textContent = teams.data ? teams.data.length : 0;
        document.getElementById('totalPlayers').textContent = players.data ? players.data.length : 0;
        document.getElementById('totalMatches').textContent = matches.data ? matches.data.length : 0;

    } catch (error) {
        console.error('Failed to load dashboard data:', error);
        showError('ไม่สามารถโหลดข้อมูลได้');
    }
}

function logout() {
    const sessionId = localStorage.getItem('sessionId');
    
    if (sessionId) {
        fetch(`${API_BASE_URL}/api/auth/logout`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${sessionId}`
            }
        }).catch(error => {
            console.error('Logout request failed:', error);
        });
    }

    localStorage.removeItem('sessionId');
    localStorage.removeItem('user');
    window.location.href = '/login.html';
}

async function testAPI(endpoint) {
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`);
        const data = await response.json();
        
        if (response.ok) {
            alert(`✅ API Test สำเร็จ!\n\nEndpoint: ${endpoint}\nจำนวนข้อมูล: ${data.data ? data.data.length : 'N/A'}`);
        } else {
            alert(`❌ API Test ล้มเหลว!\n\nEndpoint: ${endpoint}\nError: ${data.error || 'Unknown error'}`);
        }
    } catch (error) {
        alert(`❌ API Test ล้มเหลว!\n\nEndpoint: ${endpoint}\nError: ${error.message}`);
    }
}

function showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}
