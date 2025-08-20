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
                <button class="btn" @click="addNotice">추가</button>
                <button class="btn ghost" @click="clearNotices" title="모든 공지 삭제">전체삭제</button>
            </div>

            <div class="notice">
                <span style="font-weight: 600; font-size: 1.1rem;">수행평가 및 공지사항</span>
                <ul class="notice-ul">
                    <li v-for="(notice, index) in notices" @click="removeNotice(index)" :key="index">
                        <span>{{ notice.title }}</span>
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

    try {
        const response = await axios.post('https://back-dimiboard.coder.ac/notice',
            {
                title: t,
                deadline: new Date(new Date().getTime() - (new Date().getTimezoneOffset() * 60000)).toISOString().slice(0, 10)
            }
        )
        notices.value = response.data
    } catch (error) {
        console.error('Error fetching data:', error)
    }
    await fetchNotices()
}
async function removeNotice(i){
    if (confirm('이 공지를 삭제하시겠습니까?')) {
        try {
            const response = await axios.delete('https://back-dimiboard.coder.ac/notice',
                {
                    id: notices.value[i].id,
                }
            )
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
  border: 1px solid #e89f0d;
  background: rgba(232, 159, 13, 0.17);
}

.notice-ul li span {
  margin-left: 10px;
}
</style>

