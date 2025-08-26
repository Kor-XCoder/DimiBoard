<template>
  <div class="fireworks-container" style="position:fixed; left:0;top:0; width: 100%;height: 100%;">

  </div>
  <div class="wrap" ref="wrapEl">
    <header class="topbar">
      <div class="title">
        <h1>{{ grade }}학년 {{ban}}반 인원 현황</h1>
        <span class="sub" style="color: white; font-size: 1.4rem; margin-left: 1rem;">{{ nowText }}</span>
      </div>
      <div class="toolbar">
        <input
          v-model="search"
          class="input"
          placeholder="번호 검색"
          inputmode="numeric"
          @keydown.stop
        />
        <button class="btn ghost" @click="toggleFullscreen">전체 화면</button>
        <button class="btn ok" @click="allIn">모두 교실</button>
        <button class="btn warn" @click="resetAll">초기화</button>
      </div>
    </header>

    <main class="main">
      <!-- 좌측 집계 (블록 1) -->
      <aside class="aside aside-summary">
        <div class="counter">
          <div class="card"><span class="label" style="font-size: 17px;">총원</span><span class="pill gray" style="font-size: 17px;">{{ TOTAL }}명</span></div>
          <div class="card"><span class="label" style="font-size: 17px;">현원(교실)</span><span class="pill green" style="font-size: 17px;">{{ present }}명</span></div>
          <div class="card"><span class="label" style="font-size: 17px;">결원</span><span class="pill red" style="font-size: 17px;">{{ absent }}명</span></div>
        </div>
      </aside>

      <!-- 우측 보드: 2행을 모두 차지 -->
      <section class="section">
        <div class="board">
          <div class="lane room"
               data-lane="room"
               @dragover.prevent="onLaneDragOver('room', $event)"
               @dragleave="onLaneDragLeave('room')"
               @drop.prevent="onLaneDrop('room')">
            <div class="lane-header">
              <div class="lane-title"><span>교실</span></div>
              <div class="lane-count"><span>{{ lanes.room.length }}명</span></div>
            </div>
            <div class="grid">
              <button v-for="n in sorted(lanes.room)" :key="`room-${n}`"
                      class="chip"
                      :class="{ highlight: searchNum === n }"
                      draggable="true"
                      @dragstart="onChipDragStart(n, $event)"
                      @dragend="onChipDragEnd"
                      @click="openMenu($event, n)"
                      @pointerdown="onChipPointerDown(n, $event)">
                {{ n }}
              </button>
            </div>
          </div>

          <!-- 나머지 레인들 -->
          <template v-for="lane in otherLanes" :key="lane">
            <div class="lane"
                 :data-lane="lane"
                 @dragover.prevent="onLaneDragOver(lane, $event)"
                 @dragleave="onLaneDragLeave(lane)"
                 @drop.prevent="onLaneDrop(lane)">
              <div class="lane-header">
                <div class="lane-title"><span>{{ laneTitles[lane] }}</span></div>
                <div class="lane-count"><span>{{ lanes[lane].length }}명</span></div>
              </div>
              <div class="stack">
                <!-- 일반 레인: 기존 칩 버튼 렌더 -->
                <template v-if="lane !== 'etc'">
                  <button v-for="n in sorted(lanes[lane])" :key="`${lane}-${n}`"
                          class="chip"
                          :class="{ highlight: searchNum === n }"
                          draggable="true"
                          @dragstart="onChipDragStart(n, $event)"
                          @dragend="onChipDragEnd"
                          @click="openMenu($event, n)"
                          @pointerdown="onChipPointerDown(n, $event)">
                    {{ n }}
                  </button>
                </template>

                <!-- 기타 레인: 번호+사유 카드 렌더 -->
                <template v-else>
                  <div v-for="n in sorted(lanes.etc)" :key="`etc-${n}`"
                       class="chip chip-etc"
                       :class="{ highlight: searchNum === n }"
                       :data-num="n"
                       draggable="true"
                       @dragstart="onChipDragStart(n, $event)"
                       @dragend="onChipDragEnd"
                       @click="openMenu($event, n)"
                       @pointerdown="onChipPointerDown(n, $event)">
                    <div class="num">{{ n }}</div>
                    <div class="reason">{{ reasons[n] || '사유 없음' }}</div>
                  </div>
                </template>
              </div>
            </div>
          </template>
        </div>
      </section>

      <!-- 좌측 D‑Day & 공지사항 (블록 2) -->
      <aside class="aside aside-announcement">
        <!-- 디데이 블록 -->
        <div class="dday">
          <span>9월 모의고사</span>
          <h1>{{ ddayText }}</h1>
        </div>
        
        <!-- 공지사항 -->
        <div class="notice" v-if="!ID">
          <span style="font-weight: 600; font-size: 1.1rem;">수행평가 및 공지사항</span>
            <ul class="notice-ul">
              <li v-for="(notice, index) in notices" :key="index">
                <span>{{ notice.title }}</span>
              </li>
            </ul>
        </div>
      </aside>
    </main>
  </div>

  <!-- 컨텍스트 메뉴 -->
  <div v-if="menu.open" id="menu" class="menu"
       role="menu"
       :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
       @click.stop>
    <button v-for="lane in LANES" :key="`menu-${lane}`" @click="moveTo(menu.num!, lane)">
      →  {{ laneTitles[lane] }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import JSConfetti from 'js-confetti'
import axios from 'axios'

import { Fireworks } from 'fireworks-js'

import { useRoute } from 'vue-router';

// ---- 상수/타입
type LaneId = 'room' | 'restroom' | 'hall' | 'club' | 'etc' | 'after'
const TOTAL = 30
const LANES: LaneId[] = ['room','after','club','hall','restroom','etc']
const STORAGE_KEY = 'ystudy_board_state_v1'
const laneTitles: Record<LaneId,string> = {
  room:'교실', after:'방과후', club:'동아리', hall:'복도', restroom:'화장실/정수기', etc:'기타'
}
const ddayText = ref("");
const confetti = new JSConfetti()
const route = useRoute();
const ID = computed(() => route.params.id)

async function showConfetti() {
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
  confetti.addConfetti()
  await new Promise(resolve => setTimeout(resolve, 300));
}

let confettiTimer: number | null = null;
function scheduleConfetti() {
  if (confettiTimer) clearTimeout(confettiTimer);

  const now = new Date();
  const targetTime = new Date(now);
  targetTime.setHours(22, 50, 0, 0); // 오후 10시 50분

  // 이미 지났으면 내일로 설정
  if (now > targetTime) {
    targetTime.setDate(targetTime.getDate() + 1);
  }

  const delay = targetTime.getTime() - now.getTime();

  confettiTimer = window.setTimeout(() => {
    showConfetti();
    scheduleConfetti(); // 다음 날을 위해 다시 스케줄링
  }, delay);
}

let timer: number | undefined
const grade = ref(1);
const ban = ref(4);

// 10초 마다 실행


// ---- 상태
type BoardState = Record<LaneId, number[]>
const defaultState = (): BoardState => {
  const s: BoardState = { room:[], restroom:[], hall:[], club:[], etc:[], after: [] }
  for (let n=1;n<=TOTAL;n++) s.room.push(n)
  return s
}
const loadState = (): BoardState => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return defaultState()
    const parsed = JSON.parse(raw) as Partial<BoardState>
    const legacyOut = (parsed as any).out as number[] | undefined; if (Array.isArray(legacyOut)) { (parsed.etc ??= []).push(...legacyOut) }
    // 무결성 보정: 누락된 번호는 교실로
    const seen = new Set<number>()
    LANES.forEach(l => (parsed[l] ??= []).forEach(v => seen.add(v!)))
    for (let i=1;i<=TOTAL;i++) if (!seen.has(i)) (parsed.room as number[]).push(i)
    return parsed as BoardState
  } catch { return defaultState() }
}

let lanes = reactive<BoardState>(loadState())
onMounted(async () => {
  showConfetti()
  timer = window.setInterval(async () => {
    // 9/3까지의 남은 날짜 계산
    const targetDate = new Date('2025-09-03')
    const now = new Date()
    const diff = targetDate.getTime() - now.getTime()
    const daysLeft = Math.ceil(diff / (1000 * 60 * 60 * 24))
    ddayText.value = daysLeft > 0 ? `D-${daysLeft}` : (daysLeft === 0 ? 'D-Day' : `D+${-daysLeft}`)

    try {
      const response = await axios.get('https://back-dimiboard.coder.ac/notice')
      console.log(response.data)
      notices.value = response.data
    } catch (error) {
      console.error('Error fetching data:', error)
    }
  }, 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

// 파생값
const present = computed(() => lanes.room.length)
const absent  = computed(() => TOTAL - present.value)
const otherLanes = computed(() => LANES.filter(l => l !== 'room'))
const search = ref('')

// 공지사항 상태 + 저장
type NoticeType = {
  title: string
  type: string
  deadline: Date
}

const notices = ref<NoticeType[]>([])
const newNotice = ref('')

function addNotice(){
  const t = newNotice.value.trim()
  if (!t) return
  notices.value.unshift({ title: t, type: 'local', deadline: new Date() })
  newNotice.value = ''
}
function removeNotice(i: number){
  notices.value.splice(i,1)
}
function clearNotices(){
  if (confirm('모든 공지를 삭제하시겠습니까?')) notices.value = []
}
const searchNum = computed(() => {
  const n = Number(search.value.trim())
  return Number.isFinite(n) ? n : null
})

// ===== 기타(etc) 사유: 번호 → 사유 텍스트 =====
const REASON_KEY = 'ystudy_board_reasons_v2'
const reasons = reactive<Record<number, string>>({})

// 저장: 텍스트만 보관
watch(reasons, (v) => {
  const out: Record<number, string> = {}
  for (const [k, val] of Object.entries(v)){
    const num = Number(k)
    if (Number.isFinite(num) && typeof val === 'string') out[num] = val
  }
  localStorage.setItem(REASON_KEY, JSON.stringify(out))
}, { deep: true })

// 로드
onMounted(() => {
  try{
    const raw = localStorage.getItem(REASON_KEY)
    if (raw){
      const parsed = JSON.parse(raw) as Record<string, string>
      for (const [k, val] of Object.entries(parsed)){
        const num = Number(k)
        if (Number.isFinite(num)) reasons[num] = val
      }
    }
  }catch{}
})

// 시계
const nowText = ref('')
let clockTimer: number | null = null
onMounted(async () => {
  const fmt = new Intl.DateTimeFormat('ko-KR', { dateStyle:'full', timeStyle:'medium' })
  const tick = () => nowText.value = fmt.format(new Date())
  tick()
  clockTimer = window.setInterval(tick, 1000)
  scheduleConfetti();

  // 한국 시간(KST, UTC+9) 기준으로 날짜를 계산합니다.
  const now = new Date();
  const targetDate = new Date('2025-09-03T00:00:00+09:00'); // 목표 날짜를 KST 자정으로 설정
  const diff = targetDate.getTime() - now.getTime();
  const daysLeft = Math.ceil(diff / (1000 * 60 * 60 * 24));
  ddayText.value = daysLeft > 0 ? `D-${daysLeft}` : (daysLeft === 0 ? 'D-Day' : `D+${-daysLeft}`);

  try {
    const response = await axios.get('https://back-dimiboard.coder.ac/notice')
    console.log(response.data)
    notices.value = response.data
  } catch (error) {
    console.error('Error fetching data:', error)
  }
})
onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer)
  if (confettiTimer) clearTimeout(confettiTimer);
})

// 저장 (깊은 감시)
watch(lanes, () => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(lanes))
}, { deep: true })
const sorted = (arr: number[]) => [...arr].sort((a,b)=>a-b)

// 이동 공용
function moveTo(num: number, to: LaneId) {
  let reasonText: string | null = null
  if (to === 'etc'){
    reasonText = (prompt('기타로 이동하는 사유를 입력하세요.') || '').trim()
    if (!reasonText){
      alert('사유를 입력해야 기타로 이동할 수 있습니다.')
      return
    }
  }
  // 모든 레인에서 제거 후 대상 레인에 추가
  LANES.forEach(l => {
    const idx = lanes[l].indexOf(num)
    if (idx >= 0) lanes[l].splice(idx, 1)
  })
  lanes[to].push(num)

  // 사유 관리
  if (to === 'etc' && reasonText){
    reasons[num] = reasonText
  } else {
    if (reasons[num]) delete reasons[num]
  }
  closeMenu()
}

// 모두 교실 / 초기화
function allIn() {
  const all = Array.from(new Set(LANES.flatMap(l => lanes[l]))).sort((a,b)=>a-b)
  const s = defaultState()
  s.room = all
  Object.assign(lanes, s)
  // 사유 초기화
  for (const k of Object.keys(reasons)) delete (reasons as any)[k as any]
}
function resetAll() {
  if (confirm('모든 배치를 초기화하고 저장을 삭제하시겠습니까?')) {
    localStorage.removeItem(STORAGE_KEY)
    Object.assign(lanes, defaultState())
    // 사유 초기화 + 저장 제거
    for (const k of Object.keys(reasons)) delete (reasons as any)[k as any]
    localStorage.removeItem(REASON_KEY)
  }
}

// 드래그 상태
const dragNum = ref<number|null>(null)
const hoveringLane = ref<LaneId|null>(null)
const wrapEl = ref<HTMLElement | null>(null)
const activePointerId = ref<number | null>(null)

function onChipDragStart(num: number, e: DragEvent) {
  dragNum.value = num
  e.dataTransfer?.setData('text/plain', String(num)) // DnD 표준 API.  [oai_citation:2‡MDN Web Docs](https://developer.mozilla.org/en-US/docs/Web/API/HTML_Drag_and_Drop_API?utm_source=chatgpt.com)
  e.dataTransfer!.effectAllowed = 'move'
}
function onChipDragEnd() {
  dragNum.value = null
  hoveringLane.value = null
}
function onLaneDragOver(lane: LaneId, _e: DragEvent) {
  hoveringLane.value = lane
}
function onLaneDragLeave(_lane: LaneId) {
  hoveringLane.value = null
}
function onLaneDrop(lane: LaneId) {
  if (dragNum.value != null) moveTo(dragNum.value, lane)
  dragNum.value = null
  hoveringLane.value = null
}

// ===== Pointer Events 기반 터치/펜 드래그 (모바일 대응) =====
const pointerDragging = ref(false)
const dragGhost = ref<HTMLElement | null>(null)
const dragFromNum = ref<number | null>(null)
const hoverLaneEl = ref<HTMLElement | null>(null)

function getLaneFromPoint(x: number, y: number): LaneId | null {
  const el = document.elementFromPoint(x, y) as HTMLElement | null
  const laneEl = el?.closest('[data-lane]') as HTMLElement | null
  return (laneEl?.getAttribute('data-lane') as LaneId) || null
}
function setHoverLaneClass(laneId: LaneId | null) {
  // 기존 강조 제거
  if (hoverLaneEl.value) hoverLaneEl.value.classList.remove('drop-hint')
  if (!laneId) { hoverLaneEl.value = null; return }
  const el = document.querySelector(`[data-lane="${laneId}"]`) as HTMLElement | null
  if (el) { el.classList.add('drop-hint'); hoverLaneEl.value = el }
}

function createGhost(text: string) {
  const el = document.createElement('div')
  el.className = 'chip drag-ghost'
  el.textContent = text
  document.body.appendChild(el)
  dragGhost.value = el
}
function moveGhost(x: number, y: number) {
  if (!dragGhost.value) return
  dragGhost.value.style.transform = `translate(${x}px, ${y}px)`
}
function destroyGhost(){
  if (dragGhost.value?.parentElement) dragGhost.value.parentElement.removeChild(dragGhost.value)
  dragGhost.value = null
}

function onChipPointerDown(num: number, e: PointerEvent){
  // 데스크톱 마우스는 기존 HTML5 DnD로 처리 → 터치/펜만 커스텀
  if (e.pointerType === 'mouse') return
  if (pointerDragging.value || activePointerId.value !== null) return
  activePointerId.value = e.pointerId
  ;(e.target as HTMLElement).setPointerCapture?.(e.pointerId)
  pointerDragging.value = true
  dragFromNum.value = num
  createGhost(String(num))
  moveGhost(e.clientX, e.clientY)
  wrapEl.value?.classList.add('dragging')

  // 스크롤 방지
  e.preventDefault()
}

function onPointerMove(e: PointerEvent){
  if (e.pointerId !== activePointerId.value) return
  if (!pointerDragging.value) return
  moveGhost(e.clientX, e.clientY)
  const lane = getLaneFromPoint(e.clientX, e.clientY)
  setHoverLaneClass(lane)
}
function onPointerUp(e: PointerEvent){
  if (!pointerDragging.value || e.pointerId !== activePointerId.value) return
  const lane = getLaneFromPoint(e.clientX, e.clientY)
  if (dragFromNum.value != null && lane) moveTo(dragFromNum.value, lane)
  pointerDragging.value = false
  dragFromNum.value = null
  setHoverLaneClass(null)
  destroyGhost()
  activePointerId.value = null
  wrapEl.value?.classList.remove('dragging')
}

onMounted(() => {
  window.addEventListener('pointermove', onPointerMove, { passive: false })
  window.addEventListener('pointerup', onPointerUp, { passive: false })
  window.addEventListener('pointercancel', onPointerUp, { passive: false })
})
onUnmounted(() => {
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  window.removeEventListener('pointercancel', onPointerUp)
})

// 컨텍스트 메뉴
const menu = reactive<{ open:boolean; x:number; y:number; num:number|null }>({ open:false, x:0, y:0, num:null })
function openMenu(e: MouseEvent, num: number) {
  menu.open = true
  menu.num  = num
  // 화면 가장자리 넘어가지 않게 보정
  const w = 200, h = 210
  menu.x = Math.min(e.clientX, window.innerWidth - w - 12)
  menu.y = Math.min(e.clientY, window.innerHeight - h - 12)
}
function closeMenu() { menu.open = false; menu.num = null }
window.addEventListener('click', (ev) => {
  const target = ev.target as HTMLElement
  if (target.closest('#menu') || target.closest('.chip')) return
  closeMenu()
})

// 전체화면
function toggleFullscreen() {
  const el = document.documentElement
  if (!document.fullscreenElement) el.requestFullscreen?.()
  else document.exitFullscreen?.()
}

// 공개용
defineExpose({ TOTAL, LANES, lanes, moveTo, resetAll, allIn, present, absent, laneTitles })
</script>

<style>
@import url("https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/variable/pretendardvariable.min.css");

/* Fireworks 캔버스가 상호작용을 가로채지 않도록 */
.fireworks-container, .fireworks-container *{ pointer-events: none !important; }
.fireworks-container{ z-index: 0; }
/* 실제 UI를 위로 올림 */
.wrap{ position: relative; z-index: 1; }
/* Safari 등에서 버튼 요소 드래그 안정화 */
.chip{ -webkit-user-drag: element; }
:root{
  --bg:#0f1115; --panel:#171922; --muted:#2a2f3a; --line:#222735;
  --text:#e8ecf3; --sub:#ced3db; --accent:#5ac8fa; --shadow: 0 10px 24px rgba(0,0,0,.35), inset 0 1px 0 rgba(255,255,255,.02);
  --radius:16px; --chip:#ced1de3b; --chip-hover:#ced1de61;
  --container-max: 1920px;
}
*{box-sizing:border-box}
html,body,#app{
  height:100%;
  width: 100vw;
  display: flex;
  justify-content: center;
  margin: 0;
}

body{margin:0; background:linear-gradient(180deg,#0b0d12 0%,#0f1115 100%); color:var(--text);
     font:15px/1.45 "Pretendard Variable", Pretendard,system-ui,-apple-system,Segoe UI,Roboto,Apple SD Gothic Neo,Noto Sans KR,sans-serif}
.wrap{display:grid; grid-template-columns: minmax(260px,320px) minmax(0,1fr); gap:16px; width:95vw; padding: 0;}
.topbar{
  margin-top: 1rem;
  grid-column:1 / -1; display:flex; align-items:center; justify-content:space-between; gap:16px;
  background:var(--panel); border-radius:var(--radius); padding:14px 18px; box-shadow:var(--shadow); border:1px solid var(--line);
}
.title{display:flex; gap:14px; align-items:baseline}
.title h1{margin:0; font-size:22px; font-weight:700}
.title .sub{color:var(--sub); font-size:11px}
.toolbar{display:flex; gap:8px; flex-wrap:wrap}
.input{background:var(--panel); color:var(--text); border:1px solid var(--line); padding:10px 12px; border-radius:10px; outline:none; width:200px}
.btn{background:var(--muted); color:var(--text); border:1px solid var(--line); padding:10px 12px; border-radius:10px; cursor:pointer}
.btn:hover{background:#32394a}
.btn.ghost{background:transparent}
.btn.ok{background:#1f3c33; border-color:#2a5c4e; color:#b7ffdf}
.btn.warn{background:#3c2727; border-color:#5a3838; color:#ffd1d1}

.main{
  display:grid;
  grid-template-columns: minmax(16rem,18rem) minmax(0,1fr); /* ← 2번째 칸이 전체 가변 */
  align-items:start;
  min-height: calc(100vh - 120px);
  width: 95vw;
  grid-auto-rows: auto;
}
.aside, .section{
  background:var(--panel); border-radius:var(--radius); border:1px solid var(--line); box-shadow:var(--shadow);
  grid-auto-rows: auto;
}
.section{ padding:16px; min-width:0; grid-row: 1 / span 5; grid-column: 2; }
.aside{ padding:16px; width: 16rem; }

.aside-summary{ grid-column: 1; grid-row: 1; }
.aside-announcement{ grid-column: 1; grid-row: 2; }
/* 집계/공지사항 aside heading */
.aside-heading{ margin:12px 0 8px; font-weight:700; color:var(--sub); font-size:17px; letter-spacing:.2px }

.btn.tiny{ padding:6px 8px; font-size:12px }
.counter{display:grid; gap:10px}

.card{
  background:#1a2030;
  border:1px solid var(--line);
  border-radius:12px; 
  padding:12px; 
  display:flex; 
  align-items:center; 
  justify-content:space-between;
  height: 3.2rem;
}
.label{color:var(--sub)}

.pill{padding:6px 10px; border-radius:999px; font-weight:700; letter-spacing:.4px;}
.pill.gray{background:#2e3444}
.pill.green{background:#14372a; color:#baffd5}
.pill.red{background:#3a1f1f; color:#ffc3c3}
.meta{margin-top:12px; color:var(--sub); font-size:12px}

/* 보드 */
.board{
  display:grid;
  grid-template-columns: repeat(12, minmax(0,1fr));
  gap:12px; align-items:start;
  width:100%;
  grid-auto-flow: row dense;
}
.lane{
  background:linear-gradient(180deg,#151926 0%,#121623 100%); border:1px dashed #5c698e;
  border-radius:14px; padding:12px; min-height:220px; min-width:0;
}
.lane.room{ grid-row: 1; grid-column: 1 / span 4; }
.lane-header{display:flex; justify-content:space-between; align-items:center; margin-bottom:8px}
.lane-title span{font-weight:600; font-size: 1.2rem;}

.lane-count{color:var(--sub); font-size:1rem}
 .lane[data-lane="hall"],
 .lane[data-lane="restroom"]{
   grid-row: 2;
   grid-column: span 3;
 }
 .lane[data-lane="etc"]{
   grid-row: 2;
   grid-column: span 6;
 }

.lane[data-lane="after"],
.lane[data-lane="club"]{
   grid-row: 1;
   grid-column: span 4;
 }
.grid{ display:grid; grid-template-columns: repeat(6, minmax(0,1fr)); gap:8px }
.stack{ display:flex; flex-wrap:wrap; gap:8px }

/* 칩 */
.chip{
  display:inline-flex; align-items:center; justify-content:center;
  min-width:42px; height:42px; padding:0 10px; border-radius:12px;
  background:var(--chip); border:1px solid var(--line); cursor:grab; user-select:none; font-weight:700;
  transition:background .15s ease, transform .05s ease;
  color: white;
  font-family: "Pretendard Variable", Pretendard, system-ui, -apple-system, Segoe UI, Roboto, Apple SD Gothic Neo, Noto Sans KR, sans-serif;
  font-size: 19px;
  font-weight: 700;
  /* border: 2px solid #db207d; */
}

/* 기타용 번호+사유 카드 */
.chip-etc{
  display:flex; flex-direction:column; align-items:flex-start; justify-content:center;
  min-width:100px; height:auto; padding:10px 12px; gap:6px;
}
.chip-etc .num{ font-size:18px; font-weight:800; line-height:1; }
.chip-etc .reason{ font-size:14px; font-weight:600; line-height:1.25; opacity:.95; max-width:24ch; white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
.lane.drop-hint {
  border-color:#3b82f6;
  box-shadow:0 0 0 2px rgba(59,130,246,.25) inset;
}

.chip:hover{ background:var(--chip-hover) }
.chip.highlight {
  outline:2px solid var(--accent);
  outline-offset:2px;
  filter:drop-shadow(0 0 10px rgba(90,200,250,.5));
}

/* 터치/펜 드래그용 고스트 요소 */
.drag-ghost{
  position: fixed;
  left: 0; top: 0;
  transform: translate(-9999px, -9999px);
  z-index: 9999;
  pointer-events: none;
  opacity: .85;
}

.chip{ touch-action: none; }

.board{ touch-action: manipulation; }

.wrap.dragging, .wrap.dragging .board, .wrap.dragging .chip { touch-action: none; }

.menu{
  position:fixed; z-index:50; background:#0f1118; border:1px solid #283040; border-radius:12px; padding:6px; width:150px;
  box-shadow:0 12px 28px rgba(0,0,0,.5)
}
.menu button{ width:100%; border:0; background:transparent; color:var(--text); text-align:left; padding:10px 12px; border-radius:8px; cursor:pointer }
.menu button:hover{ background:#1c2130 }

@media (max-width: 1100px){
  .wrap{ grid-template-columns: 1fr }
  .main{ grid-template-columns: 1fr }
  .lane.room{ grid-column: 1 / -1 }
  .grid{ grid-template-columns: repeat(5, minmax(0,1fr)) }
  .lane[data-lane="hall"], .lane[data-lane="restroom"]{ grid-row: auto; }
  .lane[data-lane="etc"]{ grid-row: auto; }
  .board{ grid-template-columns: 1fr; }
  .lane{ grid-column: auto; }
  .lane[data-lane="after"], .lane[data-lane="club"]{ grid-row: auto; grid-column: auto; }
  .section{ grid-row:auto; grid-column:auto; }
  .aside-summary, .aside-announcement{ grid-row:auto; grid-column:auto; }
}

.dday {
  display: flex;
  flex-direction: column;
  align-items: start;
  justify-content: center;
  margin: 0;
  margin-left: 5px;
}

.dday span{
  font-size: 1rem;
}

.dday h1{
  margin: 0;
  font-size: 1.8rem;
}

/* 공지사항 패널 */

.notice {
  margin-top: 1rem;
  margin-left: 5px;
}

.notice-ul {
  padding: 0;
  margin: 0;
  margin-top: 5px;
}

.notice-ul li {
  display: flex;
  justify-content: start;
  align-items: center;
  list-style-type: none;
  width: 100%;
  margin-top: 10px;
  border-radius: 7px;
  height: 2.2rem;
  border: 1px solid #e89f0d;
  background: rgba(232, 159, 13, 0.17);
}

.notice-ul li span {
  margin-left: 10px;
}
</style>