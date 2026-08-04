package main

// testPageHTML is a deliberately bare-bones manual smoke-test client, not the
// real product UI (that's Faz 4, the Next.js app). It exists so the game
// server can be played end to end with real WebSocket connections before the
// frontend is built.
const testPageHTML = `<!DOCTYPE html>
<html lang="tr">
<head>
<meta charset="utf-8">
<title>Quizle - Test Sayfası</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 480px; margin: 40px auto; padding: 0 16px; }
  button { padding: 10px 16px; margin: 4px 0; cursor: pointer; }
  #choices button { display: block; width: 100%; text-align: left; }
  #log { white-space: pre-wrap; font-family: monospace; font-size: 13px; background: #f0f0f0; padding: 8px; height: 200px; overflow-y: auto; }
  .score { font-weight: bold; }
</style>
</head>
<body>
  <h1>Quizle — Manuel Test Sayfası</h1>
  <p><em>Bu, ger&ccedil;ek arayüz değil (o Faz 4'te gelecek). Sadece motoru test etmek i&ccedil;in.</em></p>

  <div id="setup">
    <button id="createBtn">Yeni Oda Oluştur</button>
    <p>Oda kodu: <input id="code" placeholder="ör. AB12CD"></p>
    <p>İsmin: <input id="name" placeholder="Adın"></p>
    <button id="joinBtn">Katıl</button>
  </div>

  <div id="game" style="display:none">
    <p>Oda kodu: <b id="roomCode"></b> <button id="startBtn">Oyunu Başlat</button></p>
    <h2 id="question">Bekleniyor...</h2>
    <div id="choices"></div>
    <p id="result"></p>
  </div>

  <h3>Log</h3>
  <div id="log"></div>

<script>
let ws;
const log = (msg) => {
  const el = document.getElementById('log');
  el.textContent += msg + "\n";
  el.scrollTop = el.scrollHeight;
};

document.getElementById('createBtn').onclick = async () => {
  const res = await fetch('/rooms', { method: 'POST' });
  const data = await res.json();
  document.getElementById('code').value = data.code;
  log('Oda oluşturuldu: ' + data.code);
};

document.getElementById('joinBtn').onclick = () => {
  const code = document.getElementById('code').value.trim();
  const name = document.getElementById('name').value.trim() || 'Oyuncu';
  if (!code) { alert('Önce bir oda kodu gir veya oluştur'); return; }

  ws = new WebSocket('ws://' + location.host + '/ws?code=' + code + '&name=' + encodeURIComponent(name));
  ws.onopen = () => log('Bağlandı');
  ws.onclose = () => log('Bağlantı kapandı');
  ws.onerror = (e) => log('Hata: ' + e);
  ws.onmessage = (evt) => {
    const msg = JSON.parse(evt.data);
    log('<- ' + msg.type + ' ' + JSON.stringify(msg.payload));
    handleMessage(msg);
  };

  document.getElementById('setup').style.display = 'none';
  document.getElementById('game').style.display = 'block';
  document.getElementById('roomCode').textContent = code;
};

document.getElementById('startBtn').onclick = () => {
  const code = document.getElementById('code').value.trim();
  fetch('/rooms/' + code + '/start', { method: 'POST' });
};

function handleMessage(msg) {
  if (msg.type === 'question_started') {
    document.getElementById('question').textContent = '';
    const choicesDiv = document.getElementById('choices');
    choicesDiv.innerHTML = '';
    document.getElementById('result').textContent = '';
    msg.payload.choices.forEach((choice, i) => {
      const btn = document.createElement('button');
      btn.textContent = choice;
      btn.onclick = () => ws.send(JSON.stringify({ type: 'submit_answer', choice: i }));
      choicesDiv.appendChild(btn);
    });
    document.getElementById('question').textContent = 'Soru: ' + msg.payload.question_id;
  }
  if (msg.type === 'answer_accepted') {
    document.getElementById('result').textContent =
      msg.payload.correct ? '✅ Doğru! Puan sıralamaya göre belli olacak.' : '❌ Yanlış.';
  }
  if (msg.type === 'question_revealed') {
    const points = Object.entries(msg.payload.points_awarded).map(([id, p]) => id + ': +' + p).join(', ');
    const scores = Object.entries(msg.payload.scores).map(([id, s]) => id + ': ' + s).join(', ');
    document.getElementById('question').textContent =
      'Doğru cevap açıklandı. Bu turda kazanılan: ' + points + ' | Toplam: ' + scores;
  }
  if (msg.type === 'game_finished') {
    const scores = Object.entries(msg.payload.final_scores).map(([id, s]) => id + ': ' + s).join(', ');
    document.getElementById('question').textContent = '🏁 Oyun bitti! ' + scores;
    document.getElementById('choices').innerHTML = '';
  }
}
</script>
</body>
</html>`
