const apiBase = "http://localhost:8080/api/v1";
let token = null;

function showJSON(id, data) {
  document.getElementById(id).textContent = JSON.stringify(data, null, 2);
}

// Register
const regForm = document.getElementById("register-form");
regForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = {
    name: regForm.name.value,
    email: regForm.email.value,
    password: regForm.password.value,
  };
  const res = await fetch(`${apiBase}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  const body = await res.json();
  showJSON("register-result", body);
});

// Login
const loginForm = document.getElementById("login-form");
loginForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = {
    email: loginForm.email.value,
    password: loginForm.password.value,
  };
  const res = await fetch(`${apiBase}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  const body = await res.json();
  showJSON("login-result", body);
  if (body.token) {
    token = body.token;
    document.getElementById("app").style.display = "block";
  }
});

// Load restaurants
const btnRest = document.getElementById("load-restaurants");
btnRest.addEventListener("click", async () => {
  const res = await fetch(`${apiBase}/restaurants`);
  const data = await res.json();
  const list = document.getElementById("restaurants");
  list.innerHTML = "";
  data.forEach((r) => {
    const li = document.createElement("li");
    li.textContent = `${r.id} - ${r.name} (${r.address})`;
    list.appendChild(li);
  });
});

// Load my reservations
const btnRes = document.getElementById("load-reservations");
btnRes.addEventListener("click", async () => {
  const res = await fetch(`${apiBase}/me/reservations`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await res.json();
  const list = document.getElementById("reservations");
  list.innerHTML = "";
  data.forEach((r) => {
    const li = document.createElement("li");
    li.textContent = `${r.id} - ${r.restaurantId} - ${r.startTime}`;
    list.appendChild(li);
  });
});

// Create reservation
const reserveForm = document.getElementById("reserve-form");
reserveForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const data = {
    start: new Date(reserveForm.start.value).toISOString(),
    end: new Date(reserveForm.end.value).toISOString(),
    guests: parseInt(reserveForm.guests.value, 10),
  };
  if (reserveForm.tableId.value) {
    data.tableId = reserveForm.tableId.value;
  }
  const rid = reserveForm.restaurantId.value;
  const res = await fetch(`${apiBase}/restaurants/${rid}/reservations`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });
  const body = await res.json();
  showJSON("reserve-result", body);
});
