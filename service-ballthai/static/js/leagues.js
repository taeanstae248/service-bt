// League Management JavaScript
const API_BASE_URL = window.location.protocol + '//' + window.location.host;
let leagues = [];
let currentLeague = null;
let deleteLeagueId = null;

document.addEventListener('DOMContentLoaded', function() {
    console.log('Leagues page DOMContentLoaded at:', new Date().toISOString());
    checkAuth();
    loadLeagues();
    initializeEventListeners();
});

function checkAuth() {
    const sessionId = localStorage.getItem('sessionId');
    console.log('checkAuth called, sessionId from localStorage:', sessionId);
    
    if (!sessionId) {
        console.log('No session ID found, redirecting to login');
        window.location.href = '/login.html';
        return;
    }

    console.log('Verifying session:', sessionId);

    // Verify session is still valid
    fetch(`${API_BASE_URL}/api/auth/verify`, {
        method: 'GET',
        headers: {
            'Authorization': `Bearer ${sessionId}`
        }
    })
    .then(response => {
        console.log('Verify response status:', response.status);
        console.log('Verify response ok:', response.ok);
        if (!response.ok) {
            console.log('Response not ok, status:', response.status);
            throw new Error(`HTTP ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        console.log('Verify response data:', data);
        if (!data.success) {
            console.log('Session verification failed, logging out');
            logout();
        } else {
            console.log('Session verified successfully');
        }
    })
    .catch(error => {
        console.error('Auth verification failed:', error);
        logout();
    });
}

function logout() {
    console.log('logout() called - clearing localStorage and redirecting to login');
    localStorage.removeItem('sessionId');
    localStorage.removeItem('user');
    window.location.href = '/login.html';
}

function initializeEventListeners() {
    // Search input enter key
    document.getElementById('searchInput').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            searchLeagues();
        }
    });

    // Modal close on outside click
    window.addEventListener('click', function(event) {
        const modal = document.getElementById('leagueModal');
        const deleteModal = document.getElementById('deleteModal');
        if (event.target === modal) {
            closeModal();
        }
        if (event.target === deleteModal) {
            closeDeleteModal();
        }
    });
}

async function loadLeagues() {
    showLoading(true);
    try {
        const response = await fetch(`${API_BASE_URL}/api/leagues`);
        const data = await response.json();
        
        if (data.success) {
            leagues = data.data || [];
            renderLeagues(leagues);
        } else {
            showAlert('เกิดข้อผิดพลาดในการโหลดข้อมูลลีก', 'error');
        }
    } catch (error) {
        console.error('Error loading leagues:', error);
        showAlert('เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    } finally {
        showLoading(false);
    }
}

async function searchLeagues() {
    const searchTerm = document.getElementById('searchInput').value.trim();
    showLoading(true);
    
    try {
        const url = searchTerm 
            ? `${API_BASE_URL}/api/leagues/search?q=${encodeURIComponent(searchTerm)}`
            : `${API_BASE_URL}/api/leagues`;
            
        const response = await fetch(url);
        const data = await response.json();
        
        if (data.success) {
            leagues = data.data || [];
            renderLeagues(leagues);
        } else {
            showAlert('เกิดข้อผิดพลาดในการค้นหา', 'error');
        }
    } catch (error) {
        console.error('Error searching leagues:', error);
        showAlert('เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    } finally {
        showLoading(false);
    }
}

function renderLeagues(leaguesList) {
    const container = document.getElementById('leaguesGrid');
    
    if (!leaguesList || leaguesList.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <h3>ไม่พบข้อมูลลีก</h3>
                <p>ยังไม่มีลีกในระบบ หรือลองเปลี่ยนคำค้นหา</p>
                <button onclick="openAddModal()" class="btn btn-primary">เพิ่มลีกใหม่</button>
            </div>
        `;
        return;
    }

    container.innerHTML = leaguesList.map(league => `
        <div class="league-card">
            <div class="league-info">
                <h3>${escapeHtml(league.name)}</h3>
                <p>ID: ${league.id}</p>
            </div>
            <div class="league-actions">
                <button onclick="editLeague(${league.id})" class="btn btn-edit">แก้ไข</button>
                <button onclick="deleteLeague(${league.id})" class="btn btn-delete">ลบ</button>
            </div>
        </div>
    `).join('');
}

function openAddModal() {
    currentLeague = null;
    document.getElementById('modalTitle').textContent = 'เพิ่มลีกใหม่';
    document.getElementById('leagueForm').reset();
    document.getElementById('leagueModal').style.display = 'block';
    document.getElementById('leagueName').focus();
}

function editLeague(id) {
    const league = leagues.find(l => l.id === id);
    if (!league) {
        showAlert('ไม่พบข้อมูลลีก', 'error');
        return;
    }

    currentLeague = league;
    document.getElementById('modalTitle').textContent = 'แก้ไขลีก';
    document.getElementById('leagueName').value = league.name;
    document.getElementById('leagueModal').style.display = 'block';
    document.getElementById('leagueName').focus();
}

function closeModal() {
    document.getElementById('leagueModal').style.display = 'none';
    currentLeague = null;
}

async function saveLeague(event) {
    event.preventDefault();
    
    const formData = new FormData(event.target);
    const leagueData = {
        name: formData.get('name').trim()
    };

    // Validation
    if (!leagueData.name) {
        showAlert('กรุณากรอกชื่อลีก', 'error');
        return;
    }

    showLoading(true);
    
    try {
        const url = currentLeague 
            ? `${API_BASE_URL}/api/leagues/${currentLeague.id}`
            : `${API_BASE_URL}/api/leagues`;
            
        const method = currentLeague ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(leagueData)
        });

        const data = await response.json();
        
        if (data.success) {
            showAlert(currentLeague ? 'แก้ไขลีกสำเร็จ' : 'เพิ่มลีกสำเร็จ', 'success');
            closeModal();
            loadLeagues(); // Reload the list
        } else {
            showAlert(data.error || 'เกิดข้อผิดพลาด', 'error');
        }
    } catch (error) {
        console.error('Error saving league:', error);
        showAlert('เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    } finally {
        showLoading(false);
    }
}

function deleteLeague(id) {
    const league = leagues.find(l => l.id === id);
    if (!league) {
        showAlert('ไม่พบข้อมูลลีก', 'error');
        return;
    }

    deleteLeagueId = id;
    document.getElementById('deleteLeagueName').textContent = league.name;
    document.getElementById('deleteModal').style.display = 'block';
}

function closeDeleteModal() {
    document.getElementById('deleteModal').style.display = 'none';
    deleteLeagueId = null;
}

async function confirmDelete() {
    if (!deleteLeagueId) return;

    showLoading(true);
    
    try {
        const response = await fetch(`${API_BASE_URL}/api/leagues/${deleteLeagueId}`, {
            method: 'DELETE'
        });

        const data = await response.json();
        
        if (data.success) {
            showAlert('ลบลีกสำเร็จ', 'success');
            closeDeleteModal();
            loadLeagues(); // Reload the list
        } else {
            showAlert(data.error || 'เกิดข้อผิดพลาดในการลบ', 'error');
        }
    } catch (error) {
        console.error('Error deleting league:', error);
        showAlert('เกิดข้อผิดพลาดในการเชื่อมต่อ', 'error');
    } finally {
        showLoading(false);
    }
}

function showLoading(show) {
    const loading = document.getElementById('loading');
    loading.style.display = show ? 'flex' : 'none';
}

function showAlert(message, type = 'success') {
    const container = document.getElementById('alertContainer');
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.innerHTML = `
        ${escapeHtml(message)}
        <button class="alert-close" onclick="this.parentElement.remove()">&times;</button>
    `;
    
    container.appendChild(alertDiv);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        if (alertDiv.parentElement) {
            alertDiv.remove();
        }
    }, 5000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
