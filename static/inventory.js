let inventory = {};
let navigationStack = [];
const app = document.getElementById('app');
let draggedBoxId = null; // for drag and drop

// Utility to fetch JSON (simulate fetch)
async function loadInventory() {
  return fetch('inventory.json').then(r => r.json());
}

function render() {
  app.innerHTML = '';
  renderBreadcrumbs();

  if (navigationStack.length === 0) {
    renderRoomList();
  } else if (navigationStack.length === 1) {
    renderFreezerList(navigationStack[0].data);
  } else if (navigationStack.length === 2) {
    renderShelfVisual(navigationStack[1].data);
  } else if (navigationStack.length === 3) {
    renderBoxSampleList(navigationStack[2].data);
  }
}
function renderBreadcrumbs() {
  if (navigationStack.length === 0) return;
  const crumb = document.createElement('div');
  crumb.className = 'breadcrumb';
  crumb.innerHTML = navigationStack
    .map((nav, idx) => {
      if (idx === navigationStack.length - 1) {
        return `<b>${nav.name}</b>`;
      } else {
        return `<a href="#" data-idx="${idx}" style="text-decoration: underline; cursor: pointer">${nav.name}</a>`;
      }
    })
    .join('<span class="sep">/</span>');
  app.appendChild(crumb);
}

// Attach the event listener ONCE, globally, after defining app:
if (!window.__breadcrumb_listener_attached) {
  document.getElementById('app').addEventListener('click', function(e) {
    // Look for breadcrumb links
    const target = e.target.closest('.breadcrumb a[data-idx]');
    if (target) {
     // alert("Breadcrumb link clicked: idx=" + target.dataset.idx); // DEBUG
      e.preventDefault();
      const idx = parseInt(target.dataset.idx, 10);
      navigationStack = navigationStack.slice(0, idx + 1);
      render();
    }
  });
  window.__breadcrumb_listener_attached = true;
}

function metaDetails(meta) {
  return Object.entries(meta)
    .map(([k, v]) => `<span class="meta">${k}: <b>${v}</b></span>`)
    .join('');
}

function renderRoomList() {
  app.innerHTML += '<h1>Rooms</h1>';
  const ul = document.createElement('ul');
  inventory.rooms.forEach(room => {
    const li = document.createElement('li');
    li.innerHTML = `<b>${room.name}</b> ${metaDetails(room.metadata)}`;
    li.onclick = () => {
      navigationStack = [{level:'rooms', id:room.id, name:room.name, data: room}];
      render();
    };
    ul.appendChild(li);
  });
  app.appendChild(ul);
}

function renderFreezerList(room) {
  app.innerHTML += `<h2>Freezers in ${room.name}</h2>`;
  const ul = document.createElement('ul');
  room.freezers.forEach(freezer => {
    const li = document.createElement('li');
    li.innerHTML = `<b>${freezer.name}</b> ${metaDetails(freezer.metadata)}`;
    li.onclick = () => {
      navigationStack.push({level:'freezers', id:freezer.id, name:freezer.name, data: freezer});
      render();
    };
    ul.appendChild(li);
  });
  app.appendChild(ul);
}

function renderShelfVisual(freezer) {
  app.innerHTML += `<h2 style="margin-top:8px;">Shelves in ${freezer.name}</h2>`;
  const shelvesDiv = document.createElement('div');
  shelvesDiv.className = 'shelves-visual';
  const shelvesSorted = freezer.shelves.slice().sort((a, b) => {
    let na = parseInt(a.name.match(/\d+/)), nb = parseInt(b.name.match(/\d+/));
    return na - nb;
  });

  shelvesSorted.forEach(shelf => {
    const shelfDiv = document.createElement('div');
    shelfDiv.className = 'shelf-visual';
    shelfDiv.dataset.shelfId = shelf.id;

    // Drag & drop handlers for shelf drop target
    shelfDiv.ondragover = (e) => {
      e.preventDefault();
      shelfDiv.classList.add('drag-over');
    };
    shelfDiv.ondragenter = (e) => {
      e.preventDefault();
      shelfDiv.classList.add('drag-over');
    };
    shelfDiv.ondragleave = (e) => {
      shelfDiv.classList.remove('drag-over');
    };
    shelfDiv.ondrop = (e) => {
      shelfDiv.classList.remove('drag-over');
      if (draggedBoxId) {
        moveBoxToShelfById(draggedBoxId, shelf.id);
        draggedBoxId = null;
        render();
      }
    };

    // Shelf label
    const lbl = document.createElement('span');
    lbl.className = 'shelf-label';
    lbl.innerText = shelf.name;
    shelfDiv.appendChild(lbl);

    // Boxes on shelf (visual)
    const boxesRow = document.createElement('div');
    boxesRow.className = 'shelf-boxes';
    (shelf.boxes || []).forEach(box => {
      const boxDiv = document.createElement('div');
      boxDiv.className = 'box-visual';
      boxDiv.draggable = true;

      boxDiv.ondragstart = (e) => {
        draggedBoxId = box.id;
        setTimeout(() => {
          boxDiv.classList.add('dragging');
        }, 0);
      };
      boxDiv.ondragend = (e) => {
        boxDiv.classList.remove('dragging');
        draggedBoxId = null;
      };

      boxDiv.onclick = (e) => {
        navigationStack.push({level:'boxes', id:box.id, name:box.name, data: box});
        render();
      };
      boxDiv.innerHTML = `
        <div class="box-name">${box.name}</div>
        <div class="box-meta">${box.metadata && box.metadata.label ? box.metadata.label : ''}</div>
        <button class="action-btn box-move-btn" onclick="event.stopPropagation();showMoveBoxDropdown('${box.id}')">Move</button>
      `;
      boxesRow.appendChild(boxDiv);
    });
    shelfDiv.appendChild(boxesRow);
    shelvesDiv.appendChild(shelfDiv);
  });
  app.appendChild(shelvesDiv);
}

// Move box to another shelf by ids (drag & drop)
function moveBoxToShelfById(boxId, toShelfId) {
  let fromShelf = null, boxObj = null, toShelf = null;
  // Find all shelves in all freezers in all rooms
  for (const room of inventory.rooms) {
    for (const freezer of room.freezers) {
      for (const shelf of freezer.shelves) {
        if (!toShelf && shelf.id === toShelfId) toShelf = shelf;
        if (!fromShelf) {
          const idx = shelf.boxes.findIndex(b => b.id === boxId);
          if (idx !== -1) {
            fromShelf = shelf;
            boxObj = shelf.boxes[idx];
          }
        }
      }
    }
  }
  if (fromShelf && toShelf && boxObj) {
    const idx = fromShelf.boxes.findIndex(b => b.id === boxId);
    if (idx !== -1) {
      fromShelf.boxes.splice(idx, 1);
      toShelf.boxes.push(boxObj);
    }
  }
}

// --- Move Box (to any shelf in any freezer in any room) ---
function showMoveBoxDropdown(boxId) {
  // Find current shelf and box
  let freezerNav = navigationStack[navigationStack.length-1];
  let freezer = freezerNav.data;
  let shelf = null;
  // Find the shelf containing the box
  for (const s of freezer.shelves) {
    if (s.boxes && s.boxes.some(b => b.id === boxId)) {
      shelf = s;
      break;
    }
  }
  const box = shelf && shelf.boxes.find(b => b.id === boxId);
  // List all shelves in all freezers in all rooms except current shelf
  const allShelves = [];
  inventory.rooms.forEach(room => {
    room.freezers.forEach(freezer => {
      freezer.shelves.forEach(shelfItem => {
        if (!(shelf && shelfItem.id === shelf.id)) {
          allShelves.push({
            id: shelfItem.id,
            name: `${room.name} / ${freezer.name} / ${shelfItem.name}`,
            shelf: shelfItem
          });
        }
      });
    });
  });

  const options = allShelves
    .map(s => `<option value="${s.id}">${s.name}</option>`)
    .join('');
  if (!options) {
    alert('No other shelves to move to.');
    return;
  }

  let moveBoxDiv = document.createElement('div');
  moveBoxDiv.innerHTML = `
    <span>Move <b>${box.name}</b> to shelf:</span>
    <select id="moveBoxSelect">${options}</select>
    <button class="action-btn" id="moveBoxGo">Go</button>
    <button class="action-btn" id="moveBoxCancel">Cancel</button>
  `;
  app.insertBefore(moveBoxDiv, app.childNodes[app.childNodes.length-1]);
  document.getElementById('moveBoxGo').onclick = () => {
    const newShelfId = document.getElementById('moveBoxSelect').value;
    const toShelfObj = allShelves.find(s => s.id === newShelfId);
    moveBoxToShelf(boxId, shelf, toShelfObj.shelf);
    render();
  };
  document.getElementById('moveBoxCancel').onclick = () => {
    moveBoxDiv.remove();
  }
}

// Move box to another shelf (anywhere, for Move button)
function moveBoxToShelf(boxId, fromShelf, toShelf) {
  const boxIdx = fromShelf.boxes.findIndex(b => b.id === boxId);
  if (boxIdx === -1) return;
  const [box] = fromShelf.boxes.splice(boxIdx, 1);
  toShelf.boxes.push(box);
}

function renderBoxSampleList(box) {
  app.innerHTML += `<h2>Samples in ${box.name}</h2>`;
  const samplesDiv = document.createElement('div');
  samplesDiv.className = 'samples-list';
  (box.samples || []).forEach(sample => {
    const sampleDiv = document.createElement('div');
    sampleDiv.className = 'sample-visual';
    sampleDiv.innerHTML = `
      <span><b>${sample.name}</b> [${sample.type}]</span>
      <span class="meta">${sample.metadata && sample.metadata.species ? sample.metadata.species : ''}</span>
      <span class="meta">${sample.metadata && sample.metadata.collected_by ? 'by ' + sample.metadata.collected_by : ''}</span>
      <button class="action-btn sample-move-btn" onclick="event.stopPropagation();showMoveSampleDropdown('${sample.id}')">Move</button>
    `;
    samplesDiv.appendChild(sampleDiv);
  });
  app.appendChild(samplesDiv);
}

function showMoveSampleDropdown(sampleId) {
  const boxNav = navigationStack[navigationStack.length-1];
  const box = boxNav.data;
  const allBoxes = [];
  inventory.rooms.forEach(room => {
    room.freezers.forEach(freezer => {
      freezer.shelves.forEach(shelf => {
        shelf.boxes.forEach(boxItem => {
          if (!(boxItem.id === box.id)) {
            allBoxes.push({
              id: boxItem.id,
              name: `${room.name} / ${freezer.name} / ${shelf.name} / ${boxItem.name}`,
              box: boxItem
            });
          }
        });
      });
    });
  });

  const options = allBoxes
    .map(b => `<option value="${b.id}">${b.name}</option>`)
    .join('');
  if (!options) {
    alert('No other boxes to move to.');
    return;
  }
  let moveSampleDiv = document.createElement('div');
  moveSampleDiv.innerHTML = `
    <span>Move sample to box:</span>
    <select id="moveSampleSelect">${options}</select>
    <button class="action-btn" id="moveSampleGo">Go</button>
    <button class="action-btn" id="moveSampleCancel">Cancel</button>
  `;
  app.insertBefore(moveSampleDiv, app.childNodes[app.childNodes.length-1]);
  document.getElementById('moveSampleGo').onclick = () => {
    const newBoxId = document.getElementById('moveSampleSelect').value;
    const toBoxObj = allBoxes.find(b => b.id === newBoxId);
    moveSampleToBox(sampleId, box, toBoxObj.box);
    render();
  };
  document.getElementById('moveSampleCancel').onclick = () => {
    moveSampleDiv.remove();
  }
}

function moveSampleToBox(sampleId, fromBox, toBox) {
  const idx = fromBox.samples.findIndex(s => s.id === sampleId);
  if (idx === -1) return;
  const [sample] = fromBox.samples.splice(idx, 1);
  toBox.samples.push(sample);
}

window.onload = async function () {
  inventory = await loadInventory();
  render();
};