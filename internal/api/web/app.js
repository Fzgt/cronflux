// Minimal dashboard: poll the REST API and render jobs and recent runs.
const REFRESH_MS = 5000;

async function getJSON(url) {
  const res = await fetch(url, { headers: { Accept: "application/json" } });
  if (!res.ok) throw new Error(`${url} -> ${res.status}`);
  return res.json();
}

function cell(text) {
  const td = document.createElement("td");
  td.textContent = text == null ? "" : String(text);
  return td;
}

function renderJobs(jobs) {
  const body = document.getElementById("jobs");
  body.replaceChildren();
  for (const j of jobs || []) {
    const tr = document.createElement("tr");
    tr.append(
      cell(j.id),
      cell(j.name),
      cell(j.spec || "—"),
      cell((j.depends_on || []).join(", ") || "—"),
      cell(j.enabled ? "yes" : "no"),
    );
    body.append(tr);
  }
}

function renderRuns(runs) {
  const body = document.getElementById("runs");
  body.replaceChildren();
  for (const r of runs || []) {
    const tr = document.createElement("tr");
    const state = cell(r.state);
    state.className = `state-${r.state}`;
    tr.append(
      cell(r.id.slice(0, 8)),
      cell(r.job_id),
      state,
      cell(r.attempt),
      cell(new Date(r.scheduled_for).toLocaleString()),
    );
    body.append(tr);
  }
}

function setStatus(ok) {
  const el = document.getElementById("status");
  el.textContent = ok ? "connected" : "disconnected";
  el.className = `pill ${ok ? "pill--ok" : "pill--bad"}`;
}

async function refresh() {
  try {
    const [jobs, runs] = await Promise.all([
      getJSON("/api/jobs"),
      getJSON("/api/runs?limit=25"),
    ]);
    renderJobs(jobs);
    renderRuns(runs);
    setStatus(true);
  } catch (err) {
    console.error(err);
    setStatus(false);
  }
}

refresh();
setInterval(refresh, REFRESH_MS);
