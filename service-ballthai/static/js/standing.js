async function fetchLeagues() {
    const select = document.getElementById('league_select');
    select.innerHTML = '<option value="">-- เลือกลีก --</option>';
    let data = null;
    try {
        const res = await fetch('/api/leagues');
        data = await res.json();
    } catch (e) {
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

// Safe attach for J-League button: call window.scrapeJLeague() if available, avoid ReferenceError
function _attachScrapeJLeagueFallback() {
    try {
        const jbtn = document.getElementById('scrapeJLeagueBtn');
        if (!jbtn) return;
        jbtn.addEventListener('click', function (e) {
            // call only when function is defined to avoid reference errors
            if (typeof window.scrapeJLeague === 'function') {
                window.scrapeJLeague();
            } else {
                // function not ready yet; give visual feedback
                jbtn.disabled = true;
                const old = jbtn.textContent;
                jbtn.textContent = 'กำลังเตรียม...';
                // wait briefly then try to call if available
                setTimeout(() => {
                    if (typeof window.scrapeJLeague === 'function') window.scrapeJLeague();
                    jbtn.disabled = false;
                    jbtn.textContent = old;
                }, 500);
            }
        });
    } catch (e) {
        // ignore
    }
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', _attachScrapeJLeagueFallback); else _attachScrapeJLeagueFallback();

let lastStageDropdown = null;
async function renderStandingsTableWithStage(standings) {
    window._debugStandings = standings;
    // 1. หา stage_name ที่มีอยู่จริงใน standings ของลีกนี้
    const stagesInStandings = {};
    if (!standings || !Array.isArray(standings)) {
        console.error('standings ไม่ถูกต้อง:', standings);
        return;
    }
    standings.forEach(s => {
        if(s.stage_name && typeof s.stage_name === 'string' && s.stage_name.trim() !== '') {
            if(!stagesInStandings[s.stage_name]) stagesInStandings[s.stage_name] = true;
        }
    });
    const stageNames = Object.keys(stagesInStandings);
    const stageZoneContainer = document.getElementById('stageZoneContainer');
    let selectedStageName = null;
    if(stageNames.length > 1) {
        let html = '<label>เลือกโซน/รอบ:</label> <select id="stage_select" class="search-input">';
        stageNames.forEach(name => {
            html += `<option value="${name}">${name}</option>`;
        });
        html += '</select>';
        stageZoneContainer.innerHTML = html;
        selectedStageName = stageNames[0];
        if(window.lastStageDropdown) window.lastStageDropdown.onchange = null;
        const dropdown = document.getElementById('stage_select');
        // set dropdown value to selectedStageName (from state) after rendering
        if (renderStandingsTableWithStage._selectedStageName) {
            selectedStageName = renderStandingsTableWithStage._selectedStageName;
            dropdown.value = selectedStageName;
        } else {
            renderStandingsTableWithStage._selectedStageName = selectedStageName;
        }
        dropdown.onchange = function() {
            renderStandingsTableWithStage._selectedStageName = this.value;
            renderStandingsTableWithStage._allStandings = standings;
            renderStandingsTableWithStage(standings);
        };
        window.lastStageDropdown = dropdown;
    } else {
        stageZoneContainer.innerHTML = '';
        renderStandingsTableWithStage._selectedStageName = null;
    }
    let filtered = standings;
    if(stageNames.length > 1 && renderStandingsTableWithStage._selectedStageName) {
        filtered = standings.filter(s => s.stage_name === renderStandingsTableWithStage._selectedStageName);
    }
    let html = `<table class="standings-table" border="1" cellpadding="4" style="width:100%;margin-top:1rem;">
        <thead><tr>
            <th>ลำดับ</th><th>ทีม</th><th>แข่ง</th><th>ชนะ</th><th>เสมอ</th><th>แพ้</th><th>ได้</th><th>เสีย</th><th>ผลต่าง</th><th>แต้ม</th><th>เลื่อน</th><th>สถานะ</th><th>จัดการ</th>
        </tr></thead><tbody>`;
    filtered.sort((a, b) => {
        // รองรับ current_rank ทั้ง Int64, int, null, undefined
        const ra = (a.current_rank && typeof a.current_rank === 'object' && a.current_rank.Int64 !== undefined)
            ? a.current_rank.Int64
            : (typeof a.current_rank === 'number' ? a.current_rank : 9999);
        const rb = (b.current_rank && typeof b.current_rank === 'object' && b.current_rank.Int64 !== undefined)
            ? b.current_rank.Int64
            : (typeof b.current_rank === 'number' ? b.current_rank : 9999);
        return ra - rb;
    });
    filtered.forEach((s,i) => {
        // determine status value robustly: support sql.NullInt64 object, number or string
        let st = null;
        if (s && s.status !== undefined && s.status !== null) {
            if (typeof s.status === 'object' && s.status.Int64 !== undefined) {
                st = Number(s.status.Int64);
            } else {
                st = Number(s.status);
            }
            if (Number.isNaN(st)) st = null;
        }
        const stText = (st === 1) ? 'OFF - ปิดการดึง' : 'ON - เปิดการดึง';
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
            <td id="status-cell-${s.id}">${stText} <button onclick="toggleStandingStatus(${s.id}, ${st === null ? 'null' : st})" style="margin-left:6px">🔁</button></td>
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
    const leagueId = parseInt(document.getElementById('league_select').value, 10);
    const rows = document.querySelectorAll('.standings-table tbody tr');
    const order = Array.from(rows).map((row,i) => ({ id: parseInt(row.dataset.id, 10), current_rank: i+1 }));
    try {
        const res = await fetch('/api/standings/order', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ league_id: leagueId, order })
        });
        const result = await res.json();
        if (result.success) {
            alert('บันทึกอันดับสำเร็จ!');
            if (typeof onLeagueChange === 'function') onLeagueChange();
        } else {
            alert('เกิดข้อผิดพลาด: ' + (result.error || JSON.stringify(result)));
        }
    } catch (err) {
        alert('เกิดข้อผิดพลาดขณะบันทึก: ' + err);
    }
}
// Modal ฟอร์มแก้ไข standings
function showEditStandingModal(standing) {
    // DEBUG: log standing object
    console.log('DEBUG standing in modal:', standing);
    // ลบ modal เดิมถ้ามี
    const oldModal = document.getElementById('editStandingModal');
    if (oldModal) oldModal.remove();
    // สร้าง modal
    const modal = document.createElement('div');
    modal.id = 'editStandingModal';
    modal.style = 'position:fixed;left:0;top:0;width:100vw;height:100vh;background:rgba(0,0,0,0.3);z-index:9999;display:flex;align-items:center;justify-content:center;';
    // compute selected status as number (support object/string/number)
    const selectedStatus = (standing && standing.status !== undefined && standing.status !== null)
        ? (typeof standing.status === 'object' && standing.status.Int64 !== undefined ? Number(standing.status.Int64) : Number(standing.status))
        : 0;
    modal.innerHTML = `
    <div style="background:#fff;padding:2rem 2.5rem;border-radius:12px;min-width:320px;max-width:95vw;box-shadow:0 2px 16px #0002;position:relative;">
        <h2 style="margin-top:0;margin-bottom:1.5rem;font-size:1.3rem;color:#667eea;">แก้ไขข้อมูลทีม: <span style='color:#222'>${standing.team_name||'-'}</span></h2>
        <form id="editStandingForm">
            <label>แข่ง: <input type="number" name="matches_played" value="${standing.matches_played||0}" min="0" required></label><br><br>
            <label>ชนะ: <input type="number" name="wins" value="${standing.wins||0}" min="0" required></label><br><br>
            <label>เสมอ: <input type="number" name="draws" value="${standing.draws||0}" min="0" required></label><br><br>
            <label>แพ้: <input type="number" name="losses" value="${standing.losses||0}" min="0" required></label><br><br>
            <label>ได้: <input type="number" name="goals_for" value="${standing.goals_for||0}" min="0" required></label><br><br>
            <label>เสีย: <input type="number" name="goals_against" value="${standing.goals_against||0}" min="0" required></label><br><br>
            <label>ผลต่าง: <input type="number" name="goal_difference" value="${standing.goal_difference||0}" required></label><br><br>
            <label>แต้ม: <input type="number" name="points" value="${standing.points||0}" min="0" required></label><br><br>
            <label>อันดับ: <input type="number" name="current_rank" value="${standing.current_rank?.Int64||1}" min="1" required></label><br><br>
            <label>สถานะ: 
                <select name="status" required>
                    <option value="0" ${selectedStatus==0?'selected':''}>ON - เปิดการดึง</option>
                    <option value="1" ${selectedStatus==1?'selected':''}>OFF - ปิดการดึง</option>
                </select>
            </label><br><br>
            <div style="text-align:right">
                <button type="button" id="cancelEditStanding">ยกเลิก</button>
                <button type="submit" style="background:#667eea;color:#fff;border:none;padding:0.5rem 1.5rem;border-radius:5px;">บันทึก</button>
            </div>
        </form>
    </div>`;
    document.body.appendChild(modal);
    // ensure select reflects computed selectedStatus (force-set to avoid template/caching mismatch)
    try {
        const selElem = modal.querySelector('select[name="status"]');
        if (selElem) {
            selElem.value = String(selectedStatus === undefined || Number.isNaN(Number(selectedStatus)) ? 0 : selectedStatus);
            console.log('[DEBUG] showEditStandingModal selectedStatus ->', selElem.value);
        }
    } catch (e) {
        console.warn('[DEBUG] failed to force-set status select', e);
    }
    document.getElementById('cancelEditStanding').onclick = () => modal.remove();
    document.getElementById('editStandingForm').onsubmit = async function(e) {
        e.preventDefault();
        // เก็บข้อมูลจากฟอร์ม
        const formData = new FormData(this);
        const data = {};
        // ฟิลด์ที่ต้องแปลงเป็น int
        const intFields = [
            'matches_played','wins','draws','losses','goals_for','goals_against','goal_difference','points','current_rank','status'
        ];
        for (const [k,v] of formData.entries()) {
            if (intFields.includes(k)) {
                data[k] = parseInt(v, 10);
            } else {
                data[k] = v;
            }
        }
        try {
            const res = await fetch(`/api/standings/${standing.id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            const result = await res.json();
            if (result.success) {
                alert('บันทึกข้อมูลสำเร็จ!');
                modal.remove();
                // อัปเดตตารางใหม่ (reload)
                if (typeof onLeagueChange === 'function') onLeagueChange();
            } else {
                alert('เกิดข้อผิดพลาด: '+(result.error||JSON.stringify(result)));
            }
        } catch (err) {
            alert('เกิดข้อผิดพลาดขณะบันทึก: '+err);
        }
    };
}

function editStanding(id) {
    // หา standing จาก window._debugStandings
    const standing = (window._debugStandings||[]).find(s => s.id==id);
    if (!standing) {
        alert('ไม่พบข้อมูลทีมนี้');
        return;
    }
    showEditStandingModal(standing);
}

// status toggling removed from UI
// toggle status (0 <-> 1) for a standing row by id
async function toggleStandingStatus(id, currentStatus) {
    // currentStatus may be null (treat as 0)
    const cur = (currentStatus === null || currentStatus === undefined) ? 0 : currentStatus;
    const newStatus = cur === 1 ? 0 : 1;
    try {
        const res = await fetch(`/api/standings/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status: newStatus })
        });
        const data = await res.json();
        if (data && data.success) {
            // update cell text
            const cell = document.getElementById(`status-cell-${id}`);
            if (cell) cell.innerHTML = (newStatus===1? 'OFF - ปิดการดึง' : 'ON - เปิดการดึง') + ` <button onclick="toggleStandingStatus(${id}, ${newStatus})" style="margin-left:6px">🔁</button>`;
            // update window._debugStandings if present
            if (window._debugStandings && Array.isArray(window._debugStandings)) {
                const s = window._debugStandings.find(x => x.id == id);
                if (s) {
                    s.status = { Int64: newStatus };
                }
            }
        } else {
            alert('ไม่สามารถเปลี่ยนสถานะ: ' + (data && data.error ? data.error : JSON.stringify(data)));
        }
    } catch (err) {
        alert('เกิดข้อผิดพลาดขณะเปลี่ยนสถานะ: ' + err);
    }
}

// เมื่อเลือกลีก ให้โหลด standings ของลีกนั้น
async function onLeagueChange() {
    const leagueId = document.getElementById('league_select').value;
    // enable/disable refresh button depending on selection
    try {
        const rbtn = document.getElementById('refreshStandingsBtn');
        if (rbtn) rbtn.disabled = !leagueId;
    } catch (e) {}
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
    // reset stage dropdown state ทุกครั้งที่เปลี่ยนลีก
    renderStandingsTableWithStage._selectedStageName = null;
    renderStandingsTableWithStage(standings);
}

// โหลดลีกและ set event handler
async function initStandingsPage() {
    await fetchLeagues();
    document.getElementById('league_select').onchange = onLeagueChange;
    // add a refresh button next to league_select to manually reload standings
    try {
        const sel = document.getElementById('league_select');
        if (sel && sel.parentNode) {
            let btn = document.getElementById('refreshStandingsBtn');
            if (!btn) {
                btn = document.createElement('button');
                btn.id = 'refreshStandingsBtn';
                btn.type = 'button';
                btn.className = 'btn-secondary';
                btn.style = 'margin-left:8px;padding:0.4rem 0.8rem;';
                btn.innerText = '🔄 รีเฟรชตาราง'; //test
                sel.parentNode.insertBefore(btn, sel.nextSibling);
            }
            btn.onclick = refreshStandings;
            // start disabled until a league is selected/populated
            try { btn.disabled = !sel.value; } catch(e){}
        }
    } catch (e) {
        console.warn('failed to attach refresh button', e);
    }
        // attach J-League scrape button handler if present
        try {
            const jbtn = document.getElementById('scrapeJLeagueBtn');
            if (jbtn) jbtn.onclick = scrapeJLeague;
        } catch (e) {}
}

// Refresh standings UI for currently selected league with a simple loading state
async function refreshStandings() {
    const btn = document.getElementById('refreshStandingsBtn');
    if (!btn) return;
    const oldText = btn.innerHTML;
    try {
        btn.disabled = true;
        btn.innerHTML = '⏳ กำลังโหลด...';
        await onLeagueChange();
    } catch (e) {
        console.error('refreshStandings error', e);
    } finally {
        btn.disabled = false;
        btn.innerHTML = oldText;
    }
}

initStandingsPage();

// Trigger scraping J-League from admin UI
function scrapeJLeague() {
    if (!confirm('ต้องการดึงข้อมูล J-League หรือไม่?')) return;
    const btn = document.getElementById('scrapeJLeagueBtn');
    if (btn) { btn.disabled = true; btn.textContent = 'กำลังดึง J-League...'; }
    const baseUrl = window.location.origin;
    fetch(baseUrl + '/scraper/jleague')
        .then(res => res.text())
        .then(text => {
            alert('ผลลัพธ์ J-League: ' + text);
            // refresh standings after scraping
            try { if (typeof onLeagueChange === 'function') onLeagueChange(); } catch(e){}
        })
        .catch(err => {
            alert('ดึง J-League ไม่สำเร็จ: ' + (err && err.message ? err.message : err));
        })
        .finally(() => {
            if (btn) { btn.disabled = false; btn.textContent = '🏟️ ดึง J-League'; }
        });
}
