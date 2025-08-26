// Teams Management JavaScript
const API_BASE_URL = window.location.protocol + '//' + window.location.host;
let teams = [];
let stadiums = [];
let currentTeam = null;

document.addEventListener('DOMContentLoaded', function() {
    console.log('Teams page DOMContentLoaded at:', new Date().toISOString());
    checkAuth();
    loadTeams();
    loadStadiums();
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

async function loadTeams() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE_URL}/api/teams`);
        const data = await response.json();
        
        if (data.success) {
            teams = data.data || [];
            renderTeams(teams);
        } else {
            showAlert('ไม่สามารถโหลดข้อมูลทีมได้', 'error');
        }
    } catch (error) {
        console.error('Error loading teams:', error);
        showAlert('เกิดข้อผิดพลาดในการโหลดข้อมูล', 'error');
    } finally {
        showLoading(false);
    }
}

async function loadStadiums() {
    try {
        const response = await fetch(`${API_BASE_URL}/api/stadiums`);
        const data = await response.json();
        
        if (data.success) {
            stadiums = data.data || [];
            populateStadiumSelect();
        }
    } catch (error) {
        console.error('Error loading stadiums:', error);
    }
}

function populateStadiumSelect() {
    const select = document.getElementById('teamStadium');
    select.innerHTML = '<option value="">เลือกสนาม</option>';
    
    stadiums.forEach(stadium => {
        const option = document.createElement('option');
        option.value = stadium.id;
        option.textContent = stadium.name;
        select.appendChild(option);
    });
}

function renderTeams(teamsToRender) {
    const container = document.getElementById('teamsContainer');
    
    if (!teamsToRender || teamsToRender.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">⚽</div>
                <h3>ไม่พบข้อมูลทีม</h3>
                <p>เริ่มต้นโดยการเพิ่มทีมใหม่</p>
            </div>
        `;
        return;
    }

    container.innerHTML = teamsToRender.map(team => {
        // รองรับ field หลายแบบจาก backend
        const teamName = team.name_th || team.name_en || team.name || team.team_name || 'undefined';
        // รองรับ field โลโก้หลายแบบ
        let logoSrc = '';
        if (team.logo_url) {
            logoSrc = team.logo_url;
        } else if (team.logo) {
            logoSrc = team.logo;
        } else if (team.logo_path) {
            logoSrc = team.logo_path;
        } else if (team.logo_filename) {
            logoSrc = '/img/teams/' + team.logo_filename;
        }
        return `
        <div class="team-card" data-team-id="${team.id}">
            <div class="team-header">
            ${logoSrc ? `<img src="${logoSrc}" alt="${teamName}" class="team-logo" onerror="this.style.display='none';">` : ''}
                <div class="team-logo-placeholder" ${logoSrc ? 'style="display:none"' : ''}>⚽</div>
                <div class="team-info">
                    <h3>${teamName}</h3>
                    <p>สนาม: ${team.stadium_name || 'ไม่ระบุ'}</p>
                    ${team.team_post_id ? `<p>Team Post ID: ${team.team_post_id}</p>` : ''}
                </div>
            </div>
            <div class="team-actions">
                <button class="btn-secondary" onclick="editTeam(${team.id})">แก้ไข</button>
                <button class="btn-secondary" onclick="uploadLogo(${team.id})">เปลี่ยนโลโก้</button>
                <button class="btn-danger" onclick="deleteTeam(${team.id}, '${teamName}')">ลบ</button>
            </div>
        </div>
        `;
    }).join('');
}

function initializeEventListeners() {
    // Search functionality
    document.getElementById('searchInput').addEventListener('input', function(e) {
        const searchTerm = e.target.value.toLowerCase();
        if (searchTerm === '') {
            renderTeams(teams);
        } else {
            const filteredTeams = teams.filter(team => 
                team.name.toLowerCase().includes(searchTerm)
            );
            renderTeams(filteredTeams);
        }
    });

    // Add team button
    document.getElementById('addTeamBtn').addEventListener('click', function() {
        currentTeam = null;
        document.getElementById('modalTitle').textContent = 'เพิ่มทีมใหม่';
        document.getElementById('teamForm').reset();
        document.getElementById('teamModal').style.display = 'block';
    });

    // Close modal buttons
    document.querySelectorAll('.close').forEach(btn => {
        btn.addEventListener('click', function() {
            this.closest('.modal').style.display = 'none';
        });
    });

    // Close modal when clicking outside
    window.addEventListener('click', function(e) {
        if (e.target.classList.contains('modal')) {
            e.target.style.display = 'none';
        }
    });

    // Team form submission
    document.getElementById('teamForm').addEventListener('submit', function(e) {
        e.preventDefault();
        saveTeam();
    });

    // Logo upload form submission
    document.getElementById('logoForm').addEventListener('submit', function(e) {
        e.preventDefault();
        uploadTeamLogo();
    });

    // File input change
    document.getElementById('logoFile').addEventListener('change', function(e) {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = function(e) {
                document.getElementById('logoPreview').src = e.target.result;
                document.getElementById('logoPreview').style.display = 'block';
            };
            reader.readAsDataURL(file);
        }
    });
}

async function saveTeam() {
    const formData = {
        name: document.getElementById('teamName').value,
        stadium_id: document.getElementById('teamStadium').value || null,
        team_post_id: document.getElementById('teamPostId').value || null
    };

    if (!formData.name) {
        showAlert('กรุณากรอกชื่อทีม', 'error');
        return;
    }

    try {
        const url = currentTeam ? 
            `${API_BASE_URL}/api/teams/${currentTeam.id}` : 
            `${API_BASE_URL}/api/teams`;
        
        const method = currentTeam ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData)
        });

        const data = await response.json();

        if (data.success) {
            showAlert(currentTeam ? 'อัพเดตทีมสำเร็จ' : 'เพิ่มทีมสำเร็จ', 'success');
            document.getElementById('teamModal').style.display = 'none';
            loadTeams();
        } else {
            showAlert(data.error || 'เกิดข้อผิดพลาด', 'error');
        }
    } catch (error) {
        console.error('Error saving team:', error);
        showAlert('เกิดข้อผิดพลาดในการบันทึก', 'error');
    }
}

function editTeam(teamId) {
    const team = teams.find(t => t.id === teamId);
    if (!team) return;

    currentTeam = team;
    document.getElementById('modalTitle').textContent = 'แก้ไขทีม';
    document.getElementById('teamName').value = team.name;
    document.getElementById('teamStadium').value = team.stadium_id || '';
    document.getElementById('teamPostId').value = team.team_post_id || '';
    // removed teamLogoUrl field
    document.getElementById('teamModal').style.display = 'block';
}

function uploadLogo(teamId) {
    const team = teams.find(t => t.id === teamId);
    if (!team) return;

    currentTeam = team;
    document.getElementById('logoModalTitle').textContent = `เปลี่ยนโลโก้ทีม: ${team.name}`;
    document.getElementById('logoForm').reset();
    document.getElementById('logoPreview').style.display = 'none';
    document.getElementById('logoModal').style.display = 'block';
}

async function uploadTeamLogo() {
    const fileInput = document.getElementById('logoFile');
    const file = fileInput.files[0];

    if (!file) {
        showAlert('กรุณาเลือกไฟล์รูปภาพ', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('logo', file);

    try {
        const response = await fetch(`${API_BASE_URL}/api/teams/${currentTeam.id}/logo`, {
            method: 'POST',
            body: formData
        });

        const data = await response.json();

        if (data.success) {
            showAlert('อัพโหลดโลโก้สำเร็จ', 'success');
            document.getElementById('logoModal').style.display = 'none';
            loadTeams();
        } else {
            showAlert(data.error || 'เกิดข้อผิดพลาดในการอัพโหลด', 'error');
        }
    } catch (error) {
        console.error('Error uploading logo:', error);
        showAlert('เกิดข้อผิดพลาดในการอัพโหลด', 'error');
    }
}

async function deleteTeam(teamId, teamName) {
    if (!confirm(`คุณต้องการลบทีม "${teamName}" หรือไม่?`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/api/teams/${teamId}`, {
            method: 'DELETE'
        });

        const data = await response.json();

        if (data.success) {
            showAlert('ลบทีมสำเร็จ', 'success');
            loadTeams();
        } else {
            showAlert(data.error || 'เกิดข้อผิดพลาดในการลบ', 'error');
        }
    } catch (error) {
        console.error('Error deleting team:', error);
        showAlert('เกิดข้อผิดพลาดในการลบ', 'error');
    }
}

function showLoading(show) {
    document.getElementById('loadingSpinner').style.display = show ? 'block' : 'none';
}

function showAlert(message, type) {
    const alertContainer = document.getElementById('alertContainer');
    const alertClass = type === 'success' ? 'alert-success' : 
                     type === 'warning' ? 'alert-warning' : 'alert-error';
    
    alertContainer.innerHTML = `
        <div class="alert ${alertClass}">
            ${message}
        </div>
    `;
    
    setTimeout(() => {
        alertContainer.innerHTML = '';
    }, 5000);
}
