<template>
  <div class="wrap">
    <header class="topbar">
      <div class="title">
        <h1>1학년 4반 인원 현황</h1>
        <span class="sub" style="color: white; font-size: 1.3rem;">{{ nowText }}</span>
      </div>
      <div class="toolbar">
        <input
          v-model="search"
          class="input"
          placeholder="번호 검색 (예: 12)"
          inputmode="numeric"
          @keydown.stop
        />
        <button class="btn ghost" @click="toggleFullscreen">전체 화면</button>
        <button class="btn ok" @click="allIn">모두 교실</button>
        <button class="btn warn" @click="resetAll">초기화</button>
      </div>
    </header>

    <main class="main">
      <!-- 좌측 집계 -->
      <aside class="aside">
        <div class="counter">
          <div class="card"><span class="label" style="font-size: 17px;">총원</span><span class="pill gray" style="font-size: 17px;">{{ TOTAL }}명</span></div>
          <div class="card"><span class="label" style="font-size: 17px;">현원(교실)</span><span class="pill green" style="font-size: 17px;">{{ present }}명</span></div>
          <div class="card"><span class="label" style="font-size: 17px;">결원</span><span class="pill red" style="font-size: 17px;">{{ absent }}명</span></div>
        </div>
      </aside>

      <!-- 우측 보드 -->
      <section class="section">
        <div class="board">
          <div class="lane room"
               data-lane="room"
               @dragover.prevent="onLaneDragOver('room', $event)"
               @dragleave="onLaneDragLeave('room')"
               @drop.prevent="onLaneDrop('room')">
            <div class="lane-header">
              <div class="lane-title">교실</div>
              <div class="lane-count">{{ lanes.room.length }}명</div>
            </div>
            <div class="grid">
              <button v-for="n in sorted(lanes.room)" :key="`room-${n}`"
                      class="chip"
                      :class="{ highlight: searchNum === n }"
                      draggable="true"
                      @dragstart="onChipDragStart(n, $event)"
                      @dragend="onChipDragEnd"
                      @click="openMenu($event, n)">
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
                <div class="lane-title">{{ laneTitles[lane] }}</div>
                <div class="lane-count">{{ lanes[lane].length }}명</div>
              </div>
              <div class="stack">
                <button v-for="n in sorted(lanes[lane])" :key="`${lane}-${n}`"
                        class="chip"
                        :class="{ highlight: searchNum === n }"
                        draggable="true"
                        @dragstart="onChipDragStart(n, $event)"
                        @dragend="onChipDragEnd"
                        @click="openMenu($event, n)">
                  {{ n }}
                </button>
              </div>
            </div>
          </template>
        </div>
      </section>
    </main>
  </div>

  <!-- 컨텍스트 메뉴 -->
  <div v-if="menu.open" id="menu" class="menu"
       role="menu"
       :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
       @click.stop>
    <button v-for="lane in LANES" :key="`menu-${lane}`" @click="moveTo(menu.num!, lane)">
      → {{ laneTitles[lane] }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'

// ---- 상수/타입
type LaneId = 'room' | 'restroom' | 'hall' | 'out' | 'club' | 'etc' | 'after'
const TOTAL = 30
const LANES: LaneId[] = ['room','after','club','hall','restroom','out','etc']
const STORAGE_KEY = 'ystudy_board_state_v1'
const laneTitles: Record<LaneId,string> = {
  room:'교실', after:'방과후', club:'동아리', hall:'복도', restroom:'화장실/정수기', out:'외출',  etc:'기타'
}

// ---- 상태
type BoardState = Record<LaneId, number[]>
const defaultState = (): BoardState => {
  const s: BoardState = { room:[], restroom:[], hall:[], out:[], club:[], etc:[], after: [] }
  for (let n=1;n<=TOTAL;n++) s.room.push(n)
  return s
}
const loadState = (): BoardState => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return defaultState()
    const parsed = JSON.parse(raw) as Partial<BoardState>
    // 무결성 보정: 누락된 번호는 교실로
    const seen = new Set<number>()
    LANES.forEach(l => (parsed[l] ??= []).forEach(v => seen.add(v!)))
    for (let i=1;i<=TOTAL;i++) if (!seen.has(i)) (parsed.room as number[]).push(i)
    return parsed as BoardState
  } catch { return defaultState() }
}
const lanes = reactive<BoardState>(loadState())

// 파생값
const present = computed(() => lanes.room.length)
const absent  = computed(() => TOTAL - present.value)
const otherLanes = computed(() => LANES.filter(l => l !== 'room'))
const search = ref('')
const searchNum = computed(() => {
  const n = Number(search.value.trim())
  return Number.isFinite(n) ? n : null
})

// 시계
const nowText = ref('')
let clockTimer: number | null = null
onMounted(() => {
  const fmt = new Intl.DateTimeFormat('ko-KR', { dateStyle:'full', timeStyle:'medium' })
  const tick = () => nowText.value = fmt.format(new Date())
  tick()
  clockTimer = window.setInterval(tick, 1000)
})
onUnmounted(() => { if (clockTimer) clearInterval(clockTimer) })

// 저장 (깊은 감시)
watch(lanes, () => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(lanes))
}, { deep: true }) // Composition API의 watch 동작.  [oai_citation:1‡vuejs.org](https://vuejs.org/guide/essentials/watchers?utm_source=chatgpt.com)

// 정렬 헬퍼
const sorted = (arr: number[]) => [...arr].sort((a,b)=>a-b)

// 이동 공용
function moveTo(num: number, to: LaneId) {
  LANES.forEach(l => {
    const idx = lanes[l].indexOf(num)
    if (idx >= 0) lanes[l].splice(idx, 1)
  })
  lanes[to].push(num)
  closeMenu()
}

// 모두 교실 / 초기화
function allIn() {
  const all = Array.from(new Set(LANES.flatMap(l => lanes[l]))).sort((a,b)=>a-b)
  const s = defaultState()
  s.room = all
  Object.assign(lanes, s)
}
function resetAll() {
  if (confirm('모든 배치를 초기화하고 저장을 삭제하시겠습니까?')) {
    localStorage.removeItem(STORAGE_KEY)
    Object.assign(lanes, defaultState())
  }
}

// 드래그 상태
const dragNum = ref<number|null>(null)
const hoveringLane = ref<LaneId|null>(null)

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

:root{
  --bg:#0f1115; --panel:#171922; --muted:#2a2f3a; --line:#222735;
  --text:#e8ecf3; --sub:#aab3c5; --accent:#5ac8fa; --shadow: 0 10px 24px rgba(0,0,0,.35), inset 0 1px 0 rgba(255,255,255,.02);
  --radius:16px; --chip:#1e2230; --chip-hover:#2a3041;
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
  grid-template-columns: minmax(260px,320px) minmax(0,1fr); /* ← 2번째 칸이 전체 가변 */
  gap:16px;
  align-items:start;
  min-height: calc(100vh - 120px);
  width: 95vw;
}
.aside, .section{
  background:var(--panel); border-radius:var(--radius); border:1px solid var(--line); box-shadow:var(--shadow);
}
.section{ padding:16px; min-width:0; } /* ← Grid 오버플로 방지 */
.aside{ padding:16px }
.counter{display:grid; gap:10px}
.card{background:#1a2030; border:1px solid var(--line); border-radius:12px; padding:12px; display:flex; align-items:center; justify-content:space-between}
.label{color:var(--sub)}
.pill{padding:6px 10px; border-radius:999px; font-weight:700; letter-spacing:.4px;}
.pill.gray{background:#2e3444}
.pill.green{background:#14372a; color:#baffd5}
.pill.red{background:#3a1f1f; color:#ffc3c3}
.meta{margin-top:12px; color:var(--sub); font-size:12px}

/* 보드 */
.board{
  display:grid;
  grid-template-columns: minmax(0,1.3fr) repeat(5, minmax(0,1fr)); /* ← 모든 칸이 화면 끝까지 확장 */
  gap:12px; align-items:start;
  width:100%;
}
.lane{
  background:linear-gradient(180deg,#151926 0%,#121623 100%); border:1px dashed #2a3041;
  border-radius:14px; padding:12px; min-height:220px; min-width:0;
}
.lane.room{ grid-column: 1 / span 2; }
.lane-header{display:flex; justify-content:space-between; align-items:center; margin-bottom:8px}
.lane-title{font-weight:700}
.lane-count{color:var(--sub); font-size:12px}
.grid{ display:grid; grid-template-columns: repeat(6, minmax(0,1fr)); gap:8px }
.stack{ display:flex; flex-wrap:wrap; gap:8px }

/* 칩 */
.chip{
  display:inline-flex; align-items:center; justify-content:center;
  min-width:42px; height:42px; padding:0 10px; border-radius:12px;
  background:var(--chip); border:1px solid var(--line); cursor:grab; user-select:none; font-weight:700;
  transition:background .15s ease, transform .05s ease
}
.chip:hover{ background:var(--chip-hover) }
.chip.highlight{ outline:2px solid var(--accent); outline-offset:2px; filter:drop-shadow(0 0 10px rgba(90,200,250,.5)) }

/* 컨텍스트 메뉴 */
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
}
</style>