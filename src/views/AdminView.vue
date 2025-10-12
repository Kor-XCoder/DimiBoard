<template>
    <div class="app">
        <h1>DimiBoard 관리자 패널</h1>

        <div class="panel">
            <div class="form">
                <input
                    v-model="newNotice"
                    class="input"
                    placeholder="공지 내용을 입력하고 Enter 또는 추가"
                    @keydown.enter.exact.prevent="addNotice"
                    @compositionstart="isComposing = true"
                    @compositionend="isComposing = false"
                />
                <select v-model="type" class="select" aria-label="과목 선택">
                    <option disabled value="">과목</option>
                    <option value="국어">국어</option>
                    <option value="영어">영어</option>
                    <option value="수학">수학</option>
                    <option value="사회">사회</option>
                    <option value="과학">과학</option>
                    <option value="Python">Python</option>
                    <option value="인공지능">인공지능</option>
                    <option value="체육">체육</option>
                    <option value="미술">미술</option>
                </select>
                <button class="btn" @click="addNotice">추가</button>
                <button class="btn ghost" @click="clearNotices" title="모든 공지 삭제">전체삭제</button>
            </div>

            <div class="notice">
                <span style="font-weight: 600; font-size: 1.1rem;">수행평가 및 공지사항</span>
                <ul class="notice-ul">
                    <li v-for="(notice, index) in notices" @click="removeNotice(index)" :key="index" :class="colorClass(notice.type)">
                        <span class="type-tag">{{ notice.type || '기타' }}</span>
                        <span class="title">{{ notice.title }}</span>
                    </li>
                </ul>
            </div>
        </div>
    </div>
</template>

<script setup lang="js">
import { ref, onMounted, watch } from 'vue'
import axios from 'axios';

const newNotice = ref('');
const notices = ref([]);
const isComposing = ref(false);
const type = ref('');

async function fetchNotices() {
  try {
    const response = await axios.get('https://back-dimiboard.coder.ac/notice')
    console.log(response.data)
    notices.value = response.data
  } catch (error) {
    console.error('Error fetching data:', error)
  }
}

onMounted(() => {
  fetchNotices()
  console.log(notices.value)
})


async function addNotice(){
    if (isComposing.value) return;
    const t = newNotice.value.trim()
    if (!t) return
    newNotice.value = ''
    const s = type.value || '기타'

    try {
        const response = await axios.post('https://back-dimiboard.coder.ac/notice',
            {
                title: t,
                type: s,
                deadline: new Date(new Date().getTime() - (new Date().getTimezoneOffset() * 60000)).toISOString().slice(0, 10)
            }
        )
        notices.value = response.data
        type.value = ''
    } catch (error) {
        console.error('Error fetching data:', error)
    }
    await fetchNotices()
}
async function removeNotice(i){
    if (confirm('이 공지를 삭제하시겠습니까?')) {
        try {
            const response = await axios.delete('https://back-dimiboard.coder.ac/notice', {
                data: { id: notices.value[i].id }
            });
            notices.value = response.data
        } catch (error) {
            console.error('Error fetching data:', error)
        }
        await fetchNotices()
    }
}
function clearNotices(){
  if (confirm('모든 공지를 삭제하시겠습니까?')) notices.value = []
}

// 과목에 따른 색상 클래스 반환 함수
function colorClass(t){
  const s = t || ''
  if (s === '국어') return 'c-orange'
  if (s === '수학') return 'c-green'
  if (s === '사회') return 'c-yellow'
  if (s === '과학') return 'c-blue'
  if (s === '영어') return 'c-red'
  if (s === 'Python' || s === '인공지능' || s === '기타') return 'c-dark'
  return ''
}
</script>

<style scoped>
.app {
    width: 100vw;
    height: 100dvh;
    display: flex;
    flex-direction: column;
    justify-content: start;
    align-items: center;
    
}

.panel{ width:min(720px, 92vw); margin-top:16px; display:flex; flex-direction:column; gap:12px }
.form{ display:flex; gap:8px; width:100% }
.input{ flex:1; min-width:0; padding:10px 12px; border-radius:10px; border:1px solid #d0d4dd; outline:none }
.select{ flex:0 0 8rem; min-width:8rem; padding:10px 12px; border-radius:10px; border:1px solid #d0d4dd; background:#fff; outline:none }
.type-tag{ display:inline-flex; align-items:center; height:1.6rem; padding:0 .5rem; border-radius:999px; font-size:.85rem; background:#f3f4f6; color:#374151; border:1px solid #d1d5db; margin-left:10px }
.btn{ padding:10px 12px; border-radius:10px; border:1px solid #d0d4dd; background:#215bcf; cursor:pointer }
.btn:hover{ filter:brightness(0.98) }
.btn.ghost{ background:transparent }
.btn.tiny{ padding:6px 8px; font-size:12px }
.list{ list-style:none; padding:0; margin:0; display:flex; flex-direction:column; gap:8px; width:100%; max-height:60vh; overflow:auto }
.item{ display:flex; align-items:center; gap:8px; padding:8px 10px; border:1px solid #e3e7ef; border-radius:10px; background:#fff }
.bullet{ width:6px; height:6px; border-radius:999px; background:#3b82f6 }
.text{ flex:1; min-width:0; white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
.empty{ color:#6b7280; font-size:13px; padding:6px 2px }


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
  border: 1px solid #e3e7ef;
}

.notice-ul li span {
  margin-left: 10px;
}

/* === Subject color variations === */
/* 국어/사회 → 노란색 */
.notice-ul li.c-yellow{ border-color:#eab308; background:rgba(234,179,8,.15); }
.notice-ul li.c-yellow .type-tag{ background:#ebda98; border-color:#f5c46b; color:#7a5d00 }

/* 수학 → 초록색 */
.notice-ul li.c-green{ border-color:#16a34a; background:rgba(22,163,74,.12); }
.notice-ul li.c-green .type-tag{ background:#dcfce7; border-color:#86efac; color:#065f46 }

/* 과학 → 파란색 */
.notice-ul li.c-blue{ border-color:#2563eb; background:rgba(37,99,235,.12); }
.notice-ul li.c-blue .type-tag{ background:#95bdf1; border-color:#93c5fd; color:#1e3a8a }

/* 미술/체육 → 보라색 */
.notice-ul li.c-purple{ border-color:#7c3aed; background:rgba(124,58,237,.12); }
.notice-ul li.c-purple .type-tag{ background:#ede9fe; border-color:#c4b5fd; color:#4c1d95 }

/* 파이썬/인공지능/기타 → 검은색(다크) */
.notice-ul li.c-black{ border-color:#111827; background:rgba(17, 24, 39, 0.738); color:#ffffff }
.notice-ul li.c-black .type-tag{ background:#111827; border-color:#374151; color:#ffffff }
.notice-ul li.c-black .title{ color:#ffffff }
/* 영어 → 빨간색 */
.notice-ul li.c-red{ border-color:#dc2626; background:rgba(220,38,38,.12); }
.notice-ul li.c-red .type-tag{ background:#f3afaf; border-color:#fecaca; color:#7f1d1d }
</style>
