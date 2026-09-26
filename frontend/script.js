/**
 * script.js
 * CloudTask frontend logic. Talks to the backend REST API whose base
 * URL is injected at container start-up into window.APP_CONFIG.API_URL
 * (see config.js + docker-entrypoint.sh). Falls back to same-origin
 * "/api" which works when the frontend is served behind the same
 * Ingress/reverse proxy as the backend.
 */

const API_URL = (window.APP_CONFIG && window.APP_CONFIG.API_URL) || '/api';

const columns = {
  todo: document.getElementById('col-todo'),
  'in-progress': document.getElementById('col-in-progress'),
  done: document.getElementById('col-done'),
};

const statusDot = document.querySelector('#apiStatus .dot');
const statusText = document.getElementById('apiStatusText');

function setApiStatus(online) {
  statusDot.classList.remove('online', 'offline');
  statusDot.classList.add(online ? 'online' : 'offline');
  statusText.textContent = online ? 'API connected' : 'API unreachable';
}

async function fetchTasks() {
  try {
    const res = await fetch(`${API_URL}/tasks`);
    if (!res.ok) throw new Error('Bad response');
    const { data } = await res.json();
    setApiStatus(true);
    renderBoard(data);
  } catch (err) {
    setApiStatus(false);
    console.error('Failed to load tasks:', err);
  }
}

function renderBoard(tasks) {
  Object.values(columns).forEach((col) => (col.innerHTML = ''));

  const grouped = { todo: [], 'in-progress': [], done: [] };
  tasks.forEach((t) => {
    if (!grouped[t.status]) grouped[t.status] = [];
    grouped[t.status].push(t);
  });

  Object.entries(grouped).forEach(([status, items]) => {
    const col = columns[status];
    if (!col) return;
    if (items.length === 0) {
      col.innerHTML = '<p class="empty-hint">No tasks</p>';
      return;
    }
    items.forEach((task) => col.appendChild(renderCard(task)));
  });
}

function renderCard(task) {
  const card = document.createElement('div');
  card.className = 'task-card';
  card.innerHTML = `
    <h3>${escapeHtml(task.title)}</h3>
    <p>${escapeHtml(task.description || '')}</p>
    <div class="task-actions">
      <select data-id="${task.id}" class="status-select">
        ${['todo', 'in-progress', 'done']
          .map((s) => `<option value="${s}" ${s === task.status ? 'selected' : ''}>${s}</option>`)
          .join('')}
      </select>
      <button data-id="${task.id}" class="delete-btn">Delete</button>
    </div>
  `;
  return card;
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

async function createTask(e) {
  e.preventDefault();
  const title = document.getElementById('taskTitle').value.trim();
  const description = document.getElementById('taskDescription').value.trim();
  const status = document.getElementById('taskStatus').value;
  if (!title) return;

  try {
    await fetch(`${API_URL}/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, description, status }),
    });
    document.getElementById('taskForm').reset();
    fetchTasks();
  } catch (err) {
    console.error('Failed to create task:', err);
  }
}

async function handleBoardClick(e) {
  if (e.target.classList.contains('delete-btn')) {
    const id = e.target.dataset.id;
    await fetch(`${API_URL}/tasks/${id}`, { method: 'DELETE' });
    fetchTasks();
  }
}

async function handleBoardChange(e) {
  if (e.target.classList.contains('status-select')) {
    const id = e.target.dataset.id;
    const status = e.target.value;
    await fetch(`${API_URL}/tasks/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status }),
    });
    fetchTasks();
  }
}

document.getElementById('taskForm').addEventListener('submit', createTask);
document.getElementById('board').addEventListener('click', handleBoardClick);
document.getElementById('board').addEventListener('change', handleBoardChange);

fetchTasks();
setInterval(fetchTasks, 8000); // light polling to reflect changes from other clients
