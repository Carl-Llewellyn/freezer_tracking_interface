// State
let currentRoom = null;
let currentFreezer = null;
let currentBox = null;
let allBoxes = [];

// Utility Functions
function showView(id) {
  document.querySelectorAll('.view').forEach(v => v.classList.add('hidden'));
  document.getElementById(id).classList.remove('hidden');
}
async function safeFetchJson(url) {
  try {
    const res = await fetch(url);
    if (!res.ok) {
      const errText = await res.text();
      console.error(`Error ${res.status} fetching ${url}: ${errText}`);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error(`Network error fetching ${url}:`, err);
    return null;
  }
}

// Login
const loginDlg = document.getElementById('loginDialog');
//loginDlg.showModal();
document.getElementById('loginSubmit').addEventListener('click', () => {
  const user = document.getElementById('usernameInput').value.trim();
  if (user) {
    sessionStorage.setItem('username', user);
    loginDlg.close();
    loadRooms();
  }
});

    loadRooms();

// Load Rooms
async function loadRooms() {
  showView('roomView');
  const rooms = await safeFetchJson('/getfreezerrooms');
  const ul = document.getElementById('roomList');
  ul.innerHTML = '';
  if (!rooms) return;
  rooms.forEach(r => {
    const li = document.createElement('li');
    const btn = document.createElement('button');
    btn.textContent = `${r.lab} – ${r.floor}`;
    btn.onclick = () => loadFreezers(r.id);
    li.append(btn);
    ul.append(li);
  });
}

// Load Freezers
async function loadFreezers(roomId) {
  currentRoom = roomId;
  showView('freezerView');
  document.getElementById('backToRooms').onclick = loadRooms;

  const freezers = await safeFetchJson(`/getfreezersinrooms?roomid=${roomId}`);
  const container = document.getElementById('freezerList');
  container.innerHTML = '';
  if (!freezers) return;
  freezers.forEach(f => {
    const card = document.createElement('div');
    card.className = 'freezer-card';
    card.innerHTML = `
      <h3>${f.name}</h3>
      <p>Model: ${f.model}</p>
      <p>Temp: ${f.current_holding_temp_c}°C</p>
      <p>Projects: ${f.manual_projects_contained}</p>
      <p>Last Calibrated: ${f.last_calibrated}</p>
    `;
    card.onclick = () => loadBoxes(f.id);
    container.append(card);
  });
}



// Load Boxes with Drag-and-Drop
async function loadBoxes(freezerId) {
  currentFreezer = freezerId;
  showView('boxView');
  document.getElementById('backToFreezers').onclick = () => loadFreezers(currentRoom);
  document.getElementById('addBoxBtn').onclick = () => document.getElementById('addBoxDialog').showModal();

  const boxes = await safeFetchJson(`/getboxesbyfreezer?freezerid=${freezerId}`) || [];
  const container = document.getElementById('shelvesContainer');
  container.innerHTML = '';

  // Create 5 shelves
  for (let i = 1; i <= 5; i++) {
    const shelf = document.createElement('div');
    shelf.className = 'shelf';
    shelf.dataset.shelf = i;
    shelf.ondragover = e => e.preventDefault();
    shelf.ondrop = e => handleDrop(e, freezerId);


    const moveAllShelfBoxesBtn = document.createElement('button');
    moveAllShelfBoxesBtn.textContent = 'Move all boxes';
    moveAllShelfBoxesBtn.title = 'Move all boxes';
    moveAllShelfBoxesBtn.onclick = async (e) => {
      e.stopPropagation();
      // 1) Load all freezers
      const freezers = await safeFetchJson('/getallfreezers') || [];
      const choiceList = freezers.map(f => `${f.id}: ${f.name}`).join('\n');
      const choice = prompt(`Select target freezer:\n${choiceList}`, String(currentFreezer));
      if (!choice) return;
      const newFreezer = choice.split(':')[0].trim();
      // 2) Ask for shelf number
      const newShelf = prompt('Enter target shelf number (1–5):', String(i));
      if (!newShelf) return;
      // 3) Call API
      const params = new URLSearchParams({
        oldshelf: String(i),
        newshelf: String(newShelf),
        oldfreezer: String(currentFreezer),
        newfreezer: newFreezer,
      });
      const res = await fetch(`/moveallboxestoshelf?${params.toString()}`);
      if (!res.ok) {
        alert(await res.text());
        return;
      }
      // 4) Refresh view
      loadBoxes(currentFreezer);
    };
    shelf.append(moveAllShelfBoxesBtn);


    boxes.filter(b => String(b.shelf) === String(i)).forEach(b => {
      const boxEl = document.createElement('div');
      boxEl.className = 'box';
      boxEl.id = `box-${b.id}`;
      boxEl.textContent = b.name;
      boxEl.draggable = true;
      boxEl.ondragstart = e => e.dataTransfer.setData('text', b.id);
      boxEl.onclick = () => loadSamples(b.id);

      // Edit button
      const editBtn = document.createElement('button');
      editBtn.textContent = '✎';
      editBtn.title = 'Edit box name';
      editBtn.onclick = (e) => {
        e.stopPropagation();
        editBox(b);
      };
      boxEl.append(editBtn);

      // Delete button
      const delBtn = document.createElement('button');
      delBtn.textContent = '🗑';
      delBtn.title = 'Delete box';
      delBtn.onclick = (e) => {
        e.stopPropagation();
        deleteBox(b);
      };
      boxEl.append(delBtn);

      shelf.append(boxEl);
    });

    container.append(shelf);
  }
}

function getOwnText(el) {
  return Array.from(el.childNodes)
    .filter(node => node.nodeType === Node.TEXT_NODE)
    .map(node => node.textContent.trim())
    .join(' ')
    .trim();
}

function editBox(box) {
  const newName = prompt('New box name:', box.name);
  if (newName && newName !== box.name) {
    fetch(`/updatebox?boxid=${box.id}&freezerid=${currentFreezer}&shelf=${box.shelf}&name=${encodeURIComponent(newName)}`)
      .then(() => loadBoxes(currentFreezer));
  }
}

function deleteBox(box) {
  if (!confirm(`Delete box "${box.name}"?`)) return;
  fetch(`/deletebox?boxid=${box.id}`).then(() => loadBoxes(currentFreezer));
}

// Handle Box Drop
async function handleDrop(e, freezerId) {
  e.preventDefault();
  const boxId = e.dataTransfer.getData('text');
  const newShelf = e.currentTarget.dataset.shelf;
  const boxEl = document.getElementById(`box-${boxId}`);
  e.currentTarget.append(boxEl);
  const name = encodeURIComponent(getOwnText(boxEl));
  await fetch(`/updatebox?boxid=${boxId}&freezerid=${freezerId}&shelf=${newShelf}&name=${name}`);
}

// Add Box Dialog
const addBoxDlg = document.getElementById('addBoxDialog');
document.getElementById('addBoxCancel').onclick = () => addBoxDlg.close();
document.getElementById('addBoxSubmit').onclick = async () => {
  const name = document.getElementById('newBoxName').value.trim();
  const shelf = document.getElementById('newBoxShelf').value;
  if (!name) return;
  const response = await fetch(`/insertbox?freezerid=${currentFreezer}&shelf=${shelf}&name=${encodeURIComponent(name)}`);
  
  if (!response.ok) {
    const errText = await response.text();
    alert(errText);
    return;
  }
  
  addBoxDlg.close();
  loadBoxes(currentFreezer);
};

// Fetch all boxes for Move
async function fetchAllBoxes() {
  allBoxes = await safeFetchJson('/getallboxes') || [];
}

// Load Samples View
async function loadSamples(boxId) {
  currentBox = boxId;
  await fetchAllBoxes();
  showView('sampleView');
  document.getElementById('backToBoxes').onclick = () => loadBoxes(currentFreezer);
  displaySamples();
}

// Display Samples with Edit/Delete/Move
async function displaySamples() {
  const ednas = await safeFetchJson(`/ednalinkbybox?boxid=${currentBox}`) || [];
  const fishes = await safeFetchJson(`/fishlinkbybox?boxid=${currentBox}`) || [];
  renderList('ednaList', ednas, 'edna');
  renderList('fishList', fishes, 'fish');
}

function renderList(listId, items, type) {
  const ul = document.getElementById(listId);
  ul.innerHTML = '';
  items.forEach(item => {
    const li = document.createElement('li');
    // Show name
    li.textContent = item.entered_name;

    // Warning if missing database ID
    const idField = type === 'fish' ? item.fish_id : item.edna_id;
    if (idField === null || idField === undefined) {
      const warn = document.createElement('span');
      warn.textContent = ' ⚠️';
      warn.title = 'Not found in database';
      warn.style.cursor = 'pointer';
      warn.onclick = () => alert('This sample was not found in the database');
      li.appendChild(warn);
    }

    // Buttons
    const editBtn = document.createElement('button'); editBtn.textContent = 'Edit';
    const delBtn = document.createElement('button'); delBtn.textContent = 'Delete';
    const moveBtn = document.createElement('button'); moveBtn.textContent = 'Move';
    editBtn.onclick = () => editSample(item.entered_name, type);
    delBtn.onclick = () => deleteSample(item.entered_name, type);
    moveBtn.onclick = () => moveSample(item.entered_name, type);
    li.append(' ', editBtn, ' ', delBtn, ' ', moveBtn);
    ul.append(li);
  });
}

// Edit Sample
function editSample(oldName, type) {
  const promptMsg = `New ${type} ID for "${oldName}" :`;
  const newName = prompt(promptMsg, oldName);
  if (newName && newName !== oldName) {
    fetch(`/update${type}link?boxid=${currentBox}&enteredname=${encodeURIComponent(oldName)}&newenteredname=${encodeURIComponent(newName)}`)
      .then(() => displaySamples());
  }
}

// Delete Sample
function deleteSample(name, type) {
  if (!confirm(`Delete ${type} "${name}"?`)) return;
  const endpoint = type === 'fish' ? '/deletefishlink' : '/deleteednalink';
  fetch(`${endpoint}?enteredname=${encodeURIComponent(name)}`)
    .then(() => displaySamples());
}

// Move Sample
function moveSample(name, type) {
  const choices = allBoxes.map(b => `${b.box_id}: ${b.lab}-Floor ${b.floor}-${b.freezer_name}-Shelf ${b.shelf}`);
  const choiceStr = choices.join('\n');
  const input = prompt(`Choose new box_id:\n${choiceStr}`, allBoxes[0]?.box_id || '');
  if (input) {
    fetch(`/update${type}link?boxid=${encodeURIComponent(input)}&enteredname=${encodeURIComponent(name)}`)
      .then(() => displaySamples());
  }
}

// Add Sample Form
const sampleForm = document.getElementById('addSampleForm');
sampleForm.addEventListener('submit', async e => {
  e.preventDefault();
  const ednaVal = document.getElementById('ednaInput').value.trim();
  const fishVal = document.getElementById('fishInput').value.trim();
  const msg = document.getElementById('message'); msg.textContent = '';
  let type, name;
  if (ednaVal && !fishVal) { type = 'edna'; name = ednaVal; }
  else if (fishVal && !ednaVal) { type = 'fish'; name = fishVal; }
  else { msg.textContent = 'Please enter exactly one ID.'; return; }
  const checkUrl = type === 'fish'
    ? `/checkfishalreadyinbox?fishid=${encodeURIComponent(name)}&boxid=${currentBox}`
    : `/checkednaalreadyinbox?ednaid=${encodeURIComponent(name)}&boxid=${currentBox}`;
  const text = await (await fetch(checkUrl)).text();
  if (text.trim() === '0') {
    const insertUrl = type === 'fish'
      ? `/insertfishlink?boxid=${currentBox}&enteredname=${encodeURIComponent(name)}`
      : `/insertednalink?boxid=${currentBox}&enteredname=${encodeURIComponent(name)}`;
    await fetch(insertUrl);
    sampleForm.reset(); displaySamples(); msg.textContent = 'Added sample.';
  } else {
    msg.textContent = text;
  }
});