async function apiPost(uri, payload) {
  const res = await fetch('/' + uri, {
    method: 'POST',
    body: JSON.stringify(payload),
  });

  return res.json();
}

async function apiGet(uri) {
  const res = await fetch('/' + uri, {
    method: 'GET',
  });

  return res.json();
}

function apiCreatePeerConnection(offer) {
  return apiPost('pc', { offer });
}

function apiCreateRoom() {
  return apiPost('room', {});
}

function apiSwitch(pcId, roomId) {
  return apiPost('switch', { peerConnectionId: pcId, roomId });
}

function apiListRooms() {
  return apiGet('rooms');
}

async function createPeerConnection() {
  const pc = new RTCPeerConnection({
    iceServers: [
      // { urls: 'stun:stun.l.google.com:19302' }
    ],
  });

  pc.addEventListener('icegatheringstatechange', async () => {
    console.log('iceGatheringState', pc.iceGatheringState);
    if (pc.iceGatheringState === 'complete') {
    }
  });

  pc.addEventListener('track', (event) => {
    if (event.track.kind === 'audio') return;
    document.getElementById('video').srcObject = event.streams[0];
    console.log('track attached');
  });

  const offer = await pc.createOffer({
    offerToReceiveAudio: false,
    offerToReceiveVideo: true,
  });

  await pc.setLocalDescription(offer);

  const { id, answer } = await apiCreatePeerConnection(offer);

  await pc.setRemoteDescription(answer);

  return id;
}

window.onload = function() {
  let pcId = null;
  refreshRoomsList(pcId);

  document.getElementById('connectBtn').addEventListener('click', () => {
    createPeerConnection().then((_pc) => pcId = _pc).then(() => refreshRoomsList(pcId));
  });

  document.getElementById('createRelayBtn').addEventListener('click', () => {
    apiCreateRoom().then(() => refreshRoomsList(pcId));
  });

  document.getElementById('refreshRooms').addEventListener('click', () => {
    refreshRoomsList(pcId);
  });
};

async function refreshRoomsList(pcId) {
  const { rooms } = await apiListRooms();
  rooms.sort((a, b) => a.port - b.port);
  const p = document.getElementById('rooms-list');
  p.innerHTML = '';
  rooms.forEach(r => {
    const li = document.createElement('li');
    const text = document.createTextNode(``);
    li.appendChild(text);
    li.innerHTML = `
    ID: ${r.id}
    </br>
    Port: ${r.port}
    </br>
    Participants: ${r.participants.length}
    </br>
    `;

    if (pcId) {
      const achor = document.createElement('a');
      achor.href = `#${r.id}`;
      achor.addEventListener('click', () => {
        console.log('Switch to ', r);
        apiSwitch(pcId, r.id);
      });
      achor.innerText = `Switch`;
      li.appendChild(achor);
    }

    p.append(li);
  });
}
