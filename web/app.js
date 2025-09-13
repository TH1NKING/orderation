// Global variables
let currentToken = localStorage.getItem('authToken') || '';
let currentUser = JSON.parse(localStorage.getItem('currentUser') || 'null');
const API_BASE = 'http://localhost:8080/api/v1';

// Initialize app
document.addEventListener('DOMContentLoaded', function() {
    if (currentToken && currentUser) {
        showMainApp();
    }
});

// Utility functions
function showResult(elementId, message, isError = false) {
    const element = document.getElementById(elementId);
    element.textContent = message;
    element.className = 'result ' + (isError ? 'error' : 'success');
}

function clearResult(elementId) {
    document.getElementById(elementId).textContent = '';
    document.getElementById(elementId).className = 'result';
}

async function apiCall(url, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    if (currentToken) {
        headers['Authorization'] = `Bearer ${currentToken}`;
    }

    try {
        const response = await fetch(API_BASE + url, {
            ...options,
            headers
        });
        
        // Handle different response types
        let data = {};
        if (response.status !== 204) { // 204 No Content has no body
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                data = await response.json().catch(() => ({}));
            }
        }
        
        if (!response.ok) {
            throw new Error(data.error || `HTTP ${response.status}: ${response.statusText}`);
        }
        
        return data;
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}

// Auth functions
async function register(event) {
    event.preventDefault();
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData);
    
    try {
        clearResult('authResult');
        const result = await apiCall('/auth/register', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        showResult('authResult', '注册成功！请登录。');
        event.target.reset();
        showTab('login');
    } catch (error) {
        showResult('authResult', `注册失败: ${error.message}`, true);
    }
}

async function login(event) {
    event.preventDefault();
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData);
    
    try {
        clearResult('authResult');
        const result = await apiCall('/auth/login', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        currentToken = result.token;
        currentUser = result.user;
        localStorage.setItem('authToken', currentToken);
        localStorage.setItem('currentUser', JSON.stringify(currentUser));
        
        showResult('authResult', `登录成功！欢迎 ${currentUser.name}`);
        showMainApp();
    } catch (error) {
        showResult('authResult', `登录失败: ${error.message}`, true);
    }
}

function logout() {
    currentToken = '';
    currentUser = null;
    localStorage.removeItem('authToken');
    localStorage.removeItem('currentUser');
    
    document.getElementById('authSection').style.display = 'block';
    document.getElementById('mainApp').classList.add('hidden');
    document.getElementById('loginStatus').classList.add('hidden');
    
    showResult('authResult', '已退出登录');
}

function showMainApp() {
    document.getElementById('authSection').style.display = 'none';
    document.getElementById('mainApp').classList.remove('hidden');
    document.getElementById('loginStatus').classList.remove('hidden');
    document.getElementById('currentUser').textContent = currentUser.name;
    document.getElementById('currentRole').textContent = currentUser.role;
    
    console.log('Current user:', currentUser);
    console.log('User role:', currentUser.role);
    console.log('Is admin?', currentUser.role === 'admin');
    
    // Show admin functions if user is admin
    if (currentUser.role === 'admin') {
        console.log('Showing admin functions...');
        const createRestaurantTab = document.getElementById('createRestaurantTab');
        const createTableBtn = document.getElementById('createTableBtn');
        
        console.log('createRestaurantTab element:', createRestaurantTab);
        console.log('createTableBtn element:', createTableBtn);
        
        if (createRestaurantTab) {
            createRestaurantTab.style.display = 'inline-block';
            console.log('Showed createRestaurantTab');
        } else {
            console.error('createRestaurantTab element not found!');
        }
        
        if (createTableBtn) {
            createTableBtn.style.display = 'inline-block';
            console.log('Showed createTableBtn');
        } else {
            console.error('createTableBtn element not found!');
        }
    } else {
        console.log('User is not admin, hiding admin functions');
    }
    
    // Load initial data
    loadRestaurants();
}

// Tab functions
function showTab(tab) {
    // Auth tabs
    document.getElementById('loginTab').classList.toggle('hidden', tab !== 'login');
    document.getElementById('registerTab').classList.toggle('hidden', tab !== 'register');
    
    // Update tab buttons in auth section
    document.querySelectorAll('#authSection .tab-button').forEach(btn => {
        btn.classList.remove('active');
        if ((tab === 'login' && btn.textContent === '登录') || 
            (tab === 'register' && btn.textContent === '注册')) {
            btn.classList.add('active');
        }
    });
}

function showRestaurantTab(tab) {
    document.getElementById('restaurantListTab').classList.toggle('hidden', tab !== 'list');
    document.getElementById('restaurantCreateTab').classList.toggle('hidden', tab !== 'create');
    
    // Update active tab
    document.querySelector('#restaurantListTab').closest('.section').querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
        if ((tab === 'list' && btn.textContent === '浏览餐厅') || 
            (tab === 'create' && btn.textContent === '创建餐厅')) {
            btn.classList.add('active');
        }
    });
}

function showReservationTab(tab) {
    document.getElementById('reservationCreateTab').classList.toggle('hidden', tab !== 'create');
    document.getElementById('reservationMineTab').classList.toggle('hidden', tab !== 'mine');
    
    // Update active tab  
    document.querySelector('#reservationCreateTab').closest('.section').querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
        if ((tab === 'create' && btn.textContent === '创建预订') || 
            (tab === 'mine' && btn.textContent === '我的预订')) {
            btn.classList.add('active');
        }
    });
    
    if (tab === 'mine') {
        loadMyReservations();
    }
}

// Restaurant functions
async function loadRestaurants() {
    try {
        clearResult('restaurantResult');
        const restaurants = await apiCall('/restaurants');
        
        const container = document.getElementById('restaurantsList');
        if (restaurants.length === 0) {
            container.innerHTML = '<p>暂无餐厅</p>';
            return;
        }
        
        container.innerHTML = restaurants.map(restaurant => `
            <div class="item">
                <h4>🏪 ${restaurant.name}</h4>
                <p><strong>ID:</strong> ${restaurant.id}</p>
                <p><strong>地址:</strong> ${restaurant.address}</p>
                <p><strong>营业时间:</strong> ${restaurant.openTime} - ${restaurant.closeTime}</p>
                <p><strong>创建时间:</strong> ${new Date(restaurant.createdAt).toLocaleString()}</p>
                ${currentUser && currentUser.role === 'admin' ? 
                    `<button class="delete" onclick="deleteRestaurant('${restaurant.id}', '${restaurant.name}')">删除餐厅</button>` : 
                    ''
                }
            </div>
        `).join('');
        
        showResult('restaurantResult', `加载了 ${restaurants.length} 个餐厅`);
    } catch (error) {
        showResult('restaurantResult', `加载餐厅失败: ${error.message}`, true);
    }
}


async function createRestaurant(event) {
    event.preventDefault();
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData);
    
    try {
        clearResult('restaurantResult');
        const result = await apiCall('/restaurants', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        showResult('restaurantResult', `餐厅创建成功！ID: ${result.id}`);
        event.target.reset();
        loadRestaurants();
    } catch (error) {
        showResult('restaurantResult', `创建餐厅失败: ${error.message}`, true);
    }
}

async function deleteRestaurant(restaurantId, restaurantName) {
    if (!confirm(`确定要删除餐厅 "${restaurantName}" 吗？\n\n注意：删除餐厅后，相关的桌台和预订信息可能会受到影响。`)) {
        return;
    }
    
    try {
        clearResult('restaurantResult');
        await apiCall(`/restaurants/${restaurantId}`, {
            method: 'DELETE'
        });
        
        showResult('restaurantResult', `餐厅 "${restaurantName}" 已成功删除`);
        loadRestaurants();
    } catch (error) {
        showResult('restaurantResult', `删除餐厅失败: ${error.message}`, true);
    }
}

// Table functions
async function loadTables() {
    const restaurantId = document.getElementById('tableRestaurantId').value.trim();
    if (!restaurantId) {
        showResult('tablesResult', '请输入餐厅ID', true);
        return;
    }
    
    try {
        clearResult('tablesResult');
        const tables = await apiCall(`/restaurants/${restaurantId}/tables`);
        
        const container = document.getElementById('tablesList');
        if (tables.length === 0) {
            container.innerHTML = '<p>该餐厅暂无桌台</p>';
            return;
        }
        
        container.innerHTML = tables.map(table => `
            <div class="item">
                <h4>🪑 ${table.name}</h4>
                <p><strong>ID:</strong> ${table.id}</p>
                <p><strong>容量:</strong> ${table.capacity} 人</p>
                <p><strong>餐厅ID:</strong> ${table.restaurantId}</p>
            </div>
        `).join('');
        
        showResult('tablesResult', `加载了 ${tables.length} 个桌台`);
    } catch (error) {
        showResult('tablesResult', `加载桌台失败: ${error.message}`, true);
    }
}

function showCreateTable() {
    const form = document.getElementById('createTableForm');
    form.classList.toggle('hidden');
}

async function createTable(event) {
    event.preventDefault();
    const restaurantId = document.getElementById('tableRestaurantId').value.trim();
    if (!restaurantId) {
        showResult('tablesResult', '请先输入餐厅ID', true);
        return;
    }
    
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData);
    data.capacity = parseInt(data.capacity);
    
    try {
        clearResult('tablesResult');
        const result = await apiCall(`/restaurants/${restaurantId}/tables`, {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        showResult('tablesResult', `桌台创建成功！ID: ${result.id}`);
        event.target.reset();
        loadTables();
    } catch (error) {
        showResult('tablesResult', `创建桌台失败: ${error.message}`, true);
    }
}


// Reservation functions
async function createReservation(event) {
    event.preventDefault();
    const formData = new FormData(event.target);
    const data = Object.fromEntries(formData);
    data.guests = parseInt(data.guests);
    
    // Remove empty tableId
    if (!data.tableId) {
        delete data.tableId;
    }
    
    try {
        clearResult('reservationResult');
        const result = await apiCall(`/restaurants/${data.restaurantId}/reservations`, {
            method: 'POST',
            body: JSON.stringify({
                tableId: data.tableId || "",
                start: new Date(data.startTime).toISOString(),
                end: new Date(data.endTime).toISOString(),
                guests: data.guests
            })
        });
        
        showResult('reservationResult', `预订创建成功！预订ID: ${result.id}`);
        event.target.reset();
    } catch (error) {
        showResult('reservationResult', `创建预订失败: ${error.message}`, true);
    }
}

async function loadMyReservations() {
    try {
        clearResult('reservationResult');
        const reservations = await apiCall('/me/reservations');
        
        const container = document.getElementById('myReservationsList');
        if (!reservations || reservations.length === 0) {
            container.innerHTML = '<p>您暂无预订</p>';
            return;
        }
        
        container.innerHTML = reservations.map(reservation => `
            <div class="item">
                <h4>📅 预订 ${reservation.id}</h4>
                <p><strong>餐厅ID:</strong> ${reservation.restaurantId}</p>
                <p><strong>桌台ID:</strong> ${reservation.tableId}</p>
                <p><strong>时间:</strong> ${new Date(reservation.startTime).toLocaleString()} - ${new Date(reservation.endTime).toLocaleString()}</p>
                <p><strong>人数:</strong> ${reservation.guests} 人</p>
                <p><strong>状态:</strong> ${getStatusText(reservation.status)}</p>
                <p><strong>创建时间:</strong> ${new Date(reservation.createdAt).toLocaleString()}</p>
                ${reservation.status !== 'cancelled' ? 
                    `<button class="delete" onclick="cancelReservation('${reservation.id}')">取消预订</button>` : 
                    ''
                }
            </div>
        `).join('');
        
        showResult('reservationResult', `加载了 ${reservations.length} 个预订`);
    } catch (error) {
        showResult('reservationResult', `加载预订失败: ${error.message}`, true);
    }
}

async function cancelReservation(reservationId) {
    if (!confirm('确定要取消这个预订吗？')) {
        return;
    }
    
    try {
        await apiCall(`/reservations/${reservationId}`, {
            method: 'DELETE'
        });
        
        showResult('reservationResult', '预订已取消');
        loadMyReservations();
    } catch (error) {
        showResult('reservationResult', `取消预订失败: ${error.message}`, true);
    }
}

function getStatusText(status) {
    const statusMap = {
        'confirmed': '✅ 已确认',
        'pending': '⏳ 待确认', 
        'cancelled': '❌ 已取消',
        'completed': '✅ 已完成'
    };
    return statusMap[status] || status;
}

// Global error handler
window.addEventListener('unhandledrejection', function(event) {
    console.error('Unhandled promise rejection:', event.reason);
    showResult('globalResult', `发生错误: ${event.reason}`, true);
});