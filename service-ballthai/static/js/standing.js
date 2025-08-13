async function fetchLeagues() {
    const select = document.getElementById('league_select');
    select.innerHTML = '<option value="">-- เลือกลีก --</option>';
    let data = null;
    try {
        const res = await fetch('/api/leagues');
        data = await res.json();
        console.log('DEBUG: leagues api data', data);
    } catch(e) {
        console.error('API /api/leagues error:', e);
        data = { success: false, data: [] };
    }
    if(data && data.success && Array.isArray(data.data) && data.data.length > 0) {
        data.data.forEach(l => {
            select.innerHTML += `<option value="${l.id}">${l.name}</option>`;
        });
    } else {
        select.innerHTML += '<option value="">(ไม่มีข้อมูลลีก)</option>';
    }
}

let lastStageDropdown = null;
async function renderStandingsTableWithStage(standings, stagesData) {
    let stagesMap = {};
    window._debugStagesMap = stagesMap;
    window._debugStandings = standings;
    // ถ้าไม่ได้ส่ง stagesData ให้ใช้ window._lastStagesData
    if (!stagesData && window._lastStagesData) stagesData = window._lastStagesData;
    if (Array.isArray(stagesData)) {
        stagesData.forEach(s => {
            stagesMap[String(s.id)] = s.stage_name;
        });
        window._lastStagesData = stagesData;
    }
    const stages = {};
    if (!standings || !Array.isArray(standings)) {
        console.error('standings ไม่ถูกต้อง:', standings);
        return;
    }
    standings.forEach(s => {
        if(s.stage_id && s.stage_id.Valid) {
            const idStr = String(s.stage_id.Int64);
            let stageName = stagesMap[idStr];
            if(!stages[idStr]) stages[idStr] = stageName || idStr;
        }
    });
    const stageIds = Object.keys(stages);
    const stageZoneContainer = document.getElementById('stageZoneContainer');
    let selectedStageId = null;
    if(stageIds.length > 1) {
        let html = '<label>เลือกโซน/รอบ:</label> <select id="stage_select" class="search-input">';
        stageIds.forEach(id => {
            let stageName = stagesMap[String(id)] || `โซน/รอบ ${id}`;
            html += `<option value="${id}">${stageName}</option>`;
        });
        html += '</select>';
        stageZoneContainer.innerHTML = html;
        selectedStageId = stageIds[0];
        if(window.lastStageDropdown) window.lastStageDropdown.onchange = null;
        const dropdown = document.getElementById('stage_select');
        dropdown.onchange = function() {
            renderStandingsTableWithStage._selectedStageId = this.value;
            renderStandingsTableWithStage._allStandings = standings;
            // เรียกใหม่พร้อม stagesData เดิมเสมอ
            renderStandingsTableWithStage(standings, window._lastStagesData);
        };
        window.lastStageDropdown = dropdown;
        if(renderStandingsTableWithStage._selectedStageId) {
            selectedStageId = renderStandingsTableWithStage._selectedStageId;
        } else {
            renderStandingsTableWithStage._selectedStageId = selectedStageId;
        }
    } else {
        stageZoneContainer.innerHTML = '';
    }
    let filtered = standings;
    if(stageIds.length > 1 && renderStandingsTableWithStage._selectedStageId) {
        filtered = standings.filter(s => s.stage_id && s.stage_id.Valid && String(s.stage_id.Int64) === String(renderStandingsTableWithStage._selectedStageId));
    }
    let html = `<table class="standings-table" border="1" cellpadding="4" style="width:100%;margin-top:1rem;">
        <thead><tr>
            <th>ลำดับ</th><th>ทีม</th><th>แข่ง</th><th>ชนะ</th><th>เสมอ</th><th>แพ้</th><th>ได้</th><th>เสีย</th><th>ผลต่าง</th><th>แต้ม</th><th>เลื่อน</th><th>จัดการ</th>
        </tr></thead><tbody>`;
    filtered.sort((a,b) => (a.current_rank?.Int64||0)-(b.current_rank?.Int64||0));
    filtered.forEach((s,i) => {
        html += `<tr data-id="${s.id}" data-rank="${s.current_rank?.Int64||i+1}">
            <td>${s.current_rank?.Int64||i+1}</td>
            <td>${(s.team_name && typeof s.team_name === 'string') ? s.team_name : '-'}</td>
            <td>${s.matches_played}</td>
            <td>${s.wins}</td>
            <td>${s.draws}</td>
            <td>${s.losses}</td>
            <td>${s.goals_for}</td>
            <td>${s.goals_against}</td>
            <td>${s.goal_difference}</td>
            <td>${s.points}</td>
            <td>
                <button class="move-btn" onclick="moveRow(this, -1)" ${i===0?'disabled':''}>⬆️</button>
                <button class="move-btn" onclick="moveRow(this, 1)" ${i===filtered.length-1?'disabled':''}>⬇️</button>
            </td>
            <td><button onclick="editStanding(${s.id})">✏️</button></td>
        </tr>`;
    });
    html += '</tbody></table>';
    html += '<button class="btn-primary" onclick="saveOrder()">💾 บันทึกอันดับ</button>';
    document.getElementById('standingsContainer').innerHTML = html;
}

function moveRow(btn, dir) {
    const row = btn.closest('tr');
    const tbody = row.parentNode;
    const idx = Array.from(tbody.children).indexOf(row);
    if((dir === -1 && idx === 0) || (dir === 1 && idx === tbody.children.length-1)) return;
    if(dir === -1) tbody.insertBefore(row, tbody.children[idx-1]);
    else tbody.insertBefore(row.nextSibling, row);
    updateMoveBtnState(tbody);
}
function updateMoveBtnState(tbody) {
    Array.from(tbody.children).forEach((row,i,arr) => {
        row.querySelectorAll('.move-btn')[0].disabled = i===0;
        row.querySelectorAll('.move-btn')[1].disabled = i===arr.length-1;
    });
}
async function saveOrder() {
    const leagueId = document.getElementById('league_select').value;
    const rows = document.querySelectorAll('.standings-table tbody tr');
    const order = Array.from(rows).map((row,i) => ({ id: row.dataset.id, current_rank: i+1 }));
    alert('บันทึกอันดับ (mock): '+JSON.stringify(order));
}
function editStanding(id) {
    alert('ฟอร์มแก้ไขข้อมูลทีม/คะแนน (mock) id='+id);
}

// โหลด stages ทั้งหมดไว้ล่วงหน้า (cache)
let _cachedStagesData = null;
async function fetchStages() {
    if (_cachedStagesData) return _cachedStagesData;
    try {
        const res = await fetch('/api/stages');
        const data = await res.json();
        if (data && data.success && Array.isArray(data.data)) {
            _cachedStagesData = data.data;
            return _cachedStagesData;
        }
    } catch (e) {
        console.error('API /api/stages error:', e);
    }
    return [];
}

// เมื่อเลือกลีก ให้โหลด standings ของลีกนั้น
async function onLeagueChange() {
    const leagueId = document.getElementById('league_select').value;
    if (!leagueId) {
        document.getElementById('standingsContainer').innerHTML = '';
        document.getElementById('stageZoneContainer').innerHTML = '';
        return;
    }
    // fetch standings
    let standings = [];
    try {
        const res = await fetch('/api/standings?league_id=' + leagueId);
        const data = await res.json();
        if (data && data.success && Array.isArray(data.data)) {
            standings = data.data;
        }
    } catch (e) {
        console.error('API /api/standings error:', e);
    }
    // fetch stages
    const stagesData = await fetchStages();
    renderStandingsTableWithStage(standings, stagesData);
}

// โหลดลีกและ set event handler
async function initStandingsPage() {
    await fetchLeagues();
    document.getElementById('league_select').onchange = onLeagueChange;
}

initStandingsPage();
