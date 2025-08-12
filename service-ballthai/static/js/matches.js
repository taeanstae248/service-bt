function fetchMatches(dateParam) {
    let date = dateParam;
    if (!date) {
        date = document.getElementById('date').value;
    }
    let url = '/api/matches';
    if (date) {
        url += `?date=${date}`;
    }
    fetch(url)
        .then(res => res.json())
        .then(data => {
            const matchesContainer = document.getElementById('matchesContainer');
            matchesContainer.innerHTML = '';
            if (data.success && Array.isArray(data.data)) {
                // Group matches by league
                const leagueGroups = {};
                data.data.forEach(match => {
                    const league = match.league_name || 'ไม่ระบุลีก';
                    if (!leagueGroups[league]) leagueGroups[league] = [];
                    leagueGroups[league].push(match);
                });
                Object.keys(leagueGroups).forEach(league => {
                    // สร้างหัวข้อลีก
                    const leagueHeader = document.createElement('h2');
                    leagueHeader.textContent = league;
                    leagueHeader.style.marginTop = '32px';
                    matchesContainer.appendChild(leagueHeader);

                    // สร้างตารางโปรแกรมแต่ละลีก
                    const table = document.createElement('table');
                    table.className = 'matches-league-table';
                    table.innerHTML = `
                        <thead>
                            <tr>
                                <th>เวลาแข่งขัน</th>
                                <th class="home-team">ทีมเหย้า</th>
                                <th class="score-center">สกอร์</th>
                                <th>ทีมเยือน</th>
                                <th>จัดการ</th>
                            </tr>
                        </thead>
                        <tbody></tbody>
                    `;
                    const tbody = table.querySelector('tbody');
                    leagueGroups[league].forEach(match => {
                        let timeStr = '';
                        if (match.start_time) {
                            // Format start_time as HH:mm (remove seconds)
                            const timeParts = match.start_time.split(':');
                            if (timeParts.length >= 2) {
                                timeStr = `${timeParts[0]}:${timeParts[1]}`;
                            } else {
                                timeStr = match.start_time;
                            }
                        } else if (match.start_date) {
                            // fallback: แสดงเวลาจาก start_date ถ้าไม่มี start_time
                            const d = new Date(match.start_date);
                            timeStr = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
                        }
                        tbody.innerHTML += `
                            <tr>
                                <td>${timeStr}</td>
                                <td class="home-team">${match.home_team || ''}</td>
                                <td class="score-center">${match.home_score ?? ''} - ${match.away_score ?? ''}</td>
                                <td>${match.away_team || ''}</td>
                                <td class="actions">
                                    <button class="btn" onclick="editMatch(${match.id})">แก้ไข</button>
                                    <button class="btn" onclick="deleteMatch(${match.id})">ลบ</button>
                                </td>
                            </tr>
                        `;
                    });
                    matchesContainer.appendChild(table);
                });
            } else {
                matchesContainer.innerHTML = '<div style="padding:2rem;">ไม่พบข้อมูลแมทช์</div>';
            }
        });
}
function addMatch() {
    // Fetch leagues, stage_name, teams, channels for dropdowns
    fetch('/api/leagues').then(res => res.json()).then(leaguesData => {
        const leagueSelect = document.getElementById('league_select');
        leagueSelect.innerHTML = '<option value="">-- เลือกลีก --</option>';
        if (leaguesData.success && Array.isArray(leaguesData.data)) {
            leaguesData.data.forEach(lg => {
                leagueSelect.innerHTML += `<option value="${lg.id}">${lg.name}</option>`;
            });
        }
    });
    // เติม stage
    fetch('/api/stages')
        .then(res => {
            if (!res.ok) throw new Error('Network response was not ok');
            return res.json();
        })
        .then(stagesData => {
            const stageSelect = document.getElementById('stage_name_select');
            stageSelect.innerHTML = '<option value="">-- เลือกประเภทการแข่งขัน --</option>';
            stageSelect.innerHTML += '<option value="0">(ไม่ระบุประเภทการแข่งขัน)</option>';
            if (stagesData.success && Array.isArray(stagesData.data) && stagesData.data.length > 0) {
                stagesData.data.forEach(stage => {
                    if (stage.stage_name && stage.id) {
                        stageSelect.innerHTML += `<option value="${stage.id}">${stage.stage_name}</option>`;
                    }
                });
            } else {
                alert('ไม่พบข้อมูลประเภทการแข่งขัน (stage)');
            }
        })
        .catch(err => {
            alert('เกิดข้อผิดพลาดในการโหลดประเภทการแข่งขัน: ' + err.message);
        });
    // เติมช่องทีวีและช่องถ่ายทอดสด
    fetch('/api/channels').then(res => res.json()).then(channelsData => {
        const channelSelect = document.getElementById('channel_select');
        const liveChannelSelect = document.getElementById('live_channel_select');
        channelSelect.innerHTML = '<option value="">-- เลือกช่องทีวี --</option>';
        liveChannelSelect.innerHTML = '<option value="">-- เลือกช่องถ่ายทอดสด --</option>';
        if (channelsData.success && Array.isArray(channelsData.data)) {
            channelsData.data.forEach(ch => {
                if (ch.type === 'TV') {
                    channelSelect.innerHTML += `<option value="${ch.id}">${ch.name}</option>`;
                } else {
                    liveChannelSelect.innerHTML += `<option value="${ch.id}">${ch.name}</option>`;
                }
            });
        }
    });
    // เติมทีมเหย้า/ทีมเยือน
    fetch('/api/teams').then(res => res.json()).then(teamsData => {
        const homeTeamSelect = document.getElementById('home_team_select');
        const awayTeamSelect = document.getElementById('away_team_select');
        homeTeamSelect.innerHTML = '<option value="">-- เลือกทีมเหย้า --</option>';
        awayTeamSelect.innerHTML = '<option value="">-- เลือกทีมเยือน --</option>';
        if (teamsData.success && Array.isArray(teamsData.data)) {
            teamsData.data.forEach(team => {
                homeTeamSelect.innerHTML += `<option value="${team.id}">${team.name_th}</option>`;
                awayTeamSelect.innerHTML += `<option value="${team.id}">${team.name_th}</option>`;
            });
        }
    });
    // เปิด modal ทันที
    document.getElementById('addMatchModal').style.display = 'flex';
}

function closeAddMatchModal() {
    document.getElementById('addMatchModal').style.display = 'none';
}

document.addEventListener('DOMContentLoaded', function() {
    const addMatchForm = document.getElementById('addMatchForm');
    if (addMatchForm) {
        addMatchForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const formData = new FormData(addMatchForm);
            let stageValue = formData.get('stage_name');
            let stage_id = null;
            if (stageValue && stageValue !== "null" && stageValue !== "") {
                stage_id = Number(stageValue);
            } else {
                stage_id = null;
            }
            const payload = {
                league_id: Number(formData.get('league_id')),
                stage_id: stage_id,
                start_date: formData.get('start_date'),
                start_time: formData.get('start_time'),
                home_team_id: Number(formData.get('home_team_id')),
                away_team_id: Number(formData.get('away_team_id')),
                home_score: formData.get('home_score') ? Number(formData.get('home_score')) : null,
                away_score: formData.get('away_score') ? Number(formData.get('away_score')) : null,
                match_status: formData.get('match_status')
            };
            fetch('/api/matches', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            })
            .then(res => res.json())
            .then(result => {
                if (result.success) {
                    closeAddMatchModal();
                    fetchMatches();
                } else {
                    alert('เกิดข้อผิดพลาดในการเพิ่มแมทช์');
                }
            })
            .catch(() => alert('เกิดข้อผิดพลาดในการเชื่อมต่อเซิร์ฟเวอร์'));
        });
    }
});
function editMatch(id) {
    alert('ฟังก์ชันแก้ไขแมทช์ (Demo) ID: ' + id);
}
function deleteMatch(id) {
    if (confirm('ต้องการลบแมทช์นี้ใช่หรือไม่?')) {
        alert('ฟังก์ชันลบแมทช์ (Demo) ID: ' + id);
    }
}
window.onload = function() {
    // กำหนดค่า input date เป็นวันปัจจุบัน
    const dateInput = document.getElementById('date');
    const today = new Date();
    const yyyy = today.getFullYear();
    const mm = String(today.getMonth() + 1).padStart(2, '0');
    const dd = String(today.getDate()).padStart(2, '0');
    const todayStr = `${yyyy}-${mm}-${dd}`;
    dateInput.value = todayStr;
    fetchMatches(todayStr);
};
