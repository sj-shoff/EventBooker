class EventBooker {
    constructor() {
        this.baseUrl = '/api';
        this.currentUser = null;
        this.events = [];
        this.bookings = [];
        this.currentPage = this.detectPage();
        this.initialize();
    }
    detectPage() {
        const path = window.location.pathname;
        if (path.includes('admin')) return 'admin';
        return 'user';
    }
    initialize() {
        this.loadCurrentUser();
        this.setupEventListeners();
        this.loadEvents();
        this.startAutoRefresh();
        this.updateUI();
    }
    async loadCurrentUser() {
        const userId = localStorage.getItem('eventbooker_user_id');
        const userEmail = localStorage.getItem('eventbooker_user_email');
       
        if (userId && userEmail) {
            this.currentUser = { id: userId, email: userEmail };
            this.updateUserUI();
        }
    }
    updateUserUI() {
        if (this.currentUser && this.currentPage === 'user') {
            const userInfo = document.getElementById('user-info');
            if (userInfo) {
                userInfo.innerHTML = `
                    <div class="alert alert-info">
                        <i class="bi bi-person-circle me-2"></i>
                        <strong>Вы вошли как:</strong> ${this.currentUser.email}
                        <br>
                        <small>ID: ${this.currentUser.id}</small>
                    </div>
                `;
            }
           
            const bookUserId = document.getElementById('book-user-id');
            if (bookUserId) {
                bookUserId.value = this.currentUser.id;
            }
        }
    }
    setupEventListeners() {
        // Регистрация пользователя
        const registerForm = document.getElementById('register-form');
        if (registerForm) {
            registerForm.addEventListener('submit', (e) => this.handleRegister(e));
        }
        // Создание мероприятия
        const createEventForm = document.getElementById('create-event-form');
        if (createEventForm) {
            createEventForm.addEventListener('submit', (e) => this.handleCreateEvent(e));
        }
        // Бронирование
        const bookForm = document.getElementById('book-form');
        if (bookForm) {
            bookForm.addEventListener('submit', (e) => this.handleBook(e));
        }
        // Подтверждение
        const confirmForm = document.getElementById('confirm-form');
        if (confirmForm) {
            confirmForm.addEventListener('submit', (e) => this.handleConfirm(e));
        }
        // Отмена мероприятия (админ)
        const cancelEventForm = document.getElementById('cancel-event-form');
        if (cancelEventForm) {
            cancelEventForm.addEventListener('submit', (e) => this.handleCancelEvent(e));
        }
        // Поиск
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            searchInput.addEventListener('input', () => this.filterEvents());
        }
        // Фильтры бронирований
        const filterButtons = document.querySelectorAll('.booking-filter');
        filterButtons.forEach(btn => {
            btn.addEventListener('click', (e) => this.filterBookings(e.target.dataset.filter));
        });
    }
    async loadEvents() {
        try {
            this.showLoader();
            const response = await fetch(`${this.baseUrl}/events`);
            if (!response.ok) throw new Error('Ошибка загрузки мероприятий');
           
            this.events = await response.json();
            this.renderEvents();
            this.updateStatistics();
            this.hideLoader();
           
            this.showToast('Список мероприятий обновлен', 'success');
        } catch (error) {
            console.error('Error loading events:', error);
            this.showToast('Не удалось загрузить мероприятия', 'danger');
            this.hideLoader();
        }
    }
    async loadBookings() {
        if (this.currentPage !== 'admin') return;
       
        try {
            const response = await fetch(`${this.baseUrl}/bookings`);
            if (!response.ok) throw new Error('Ошибка загрузки бронирований');
            this.bookings = await response.json();
            this.renderBookings();
            this.updateStatistics();
        } catch (error) {
            console.error('Error loading bookings:', error);
            this.showToast('Ошибка загрузки бронирований', 'danger');
        }
    }
    renderEvents() {
        const container = document.getElementById('events-container');
        const tableBody = document.getElementById('events-table-body');
        const eventsList = document.getElementById('events-list');
       
        if (container) {
            this.renderEventCards(container);
        }
       
        if (tableBody) {
            this.renderEventTable(tableBody);
        }
       
        if (eventsList) {
            this.renderEventList(eventsList);
        }
    }
    renderEventCards(container) {
        if (this.events.length === 0) {
            container.innerHTML = `
                <div class="col-12 text-center py-5">
                    <i class="bi bi-calendar-x" style="font-size: 3rem; color: #6c757d;"></i>
                    <h4 class="mt-3 text-muted">Нет доступных мероприятий</h4>
                </div>
            `;
            return;
        }
        container.innerHTML = this.events.map(event => `
            <div class="col-lg-4 col-md-6 mb-4 fade-in">
                <div class="card h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h5 class="mb-0">${this.escapeHtml(event.Name)}</h5>
                        <span class="badge ${event.Available > 0 ? 'bg-success' : 'bg-danger'}">
                            ${event.Available > 0 ? 'Есть места' : 'Заполнено'}
                        </span>
                    </div>
                    <div class="card-body">
                        <div class="mb-3">
                            <small class="text-muted d-block mb-1">Дата:</small>
                            <strong>${this.formatDate(event.Date)}</strong>
                        </div>
                       
                        <div class="mb-3">
                            <small class="text-muted d-block mb-1">Места:</small>
                            <div class="d-flex align-items-center">
                                <div class="flex-grow-1 me-3">
                                    <div class="progress" style="height: 8px;">
                                        <div class="progress-bar ${this.getProgressColor(event)}"
                                             style="width: ${this.getOccupancyPercent(event)}%">
                                        </div>
                                    </div>
                                </div>
                                <small>${event.Available}/${event.TotalSeats}</small>
                            </div>
                        </div>
                       
                        <div class="mb-3">
                            <small class="text-muted d-block mb-1">Время на подтверждение:</small>
                            <span class="badge bg-info">
                                ${this.formatDuration(event.BookingTTL)}
                            </span>
                        </div>
                       
                        <div class="mb-3">
                            <small class="text-muted d-block mb-1">Оплата:</small>
                            <span class="badge ${event.RequiresPayment ? 'bg-warning' : 'bg-success'}">
                                ${event.RequiresPayment ? 'Требуется подтверждение' : 'Без оплаты'}
                            </span>
                        </div>
                       
                        <div class="mb-3">
                            <small class="text-muted d-block mb-1">ID мероприятия:</small>
                            <div class="input-group input-group-sm">
                                <input type="text" class="form-control" value="${event.ID}" readonly>
                                <button class="btn btn-outline-secondary" type="button"
                                        onclick="app.copyToClipboard('${event.ID}')">
                                    <i class="bi bi-clipboard"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    <div class="card-footer bg-transparent">
                        <button class="btn ${event.Available > 0 ? 'btn-primary' : 'btn-secondary'} w-100"
                                onclick="app.quickBook('${event.ID}')"
                                ${event.Available === 0 ? 'disabled' : ''}>
                            <i class="bi bi-bookmark-plus me-2"></i>
                            ${event.Available > 0 ? 'Забронировать' : 'Мест нет'}
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }
    renderEventTable(tableBody) {
        tableBody.innerHTML = this.events.map(event => `
            <tr>
                <td>
                    <strong>${this.escapeHtml(event.Name)}</strong><br>
                    <small class="text-muted">ID: ${event.ID}</small>
                </td>
                <td>${this.formatDate(event.Date)}</td>
                <td>
                    <div class="d-flex align-items-center">
                        <div class="progress flex-grow-1 me-2" style="height: 6px;">
                            <div class="progress-bar ${this.getProgressColor(event)}"
                                 style="width: ${this.getOccupancyPercent(event)}%">
                            </div>
                        </div>
                        <small>${event.Available}/${event.TotalSeats}</small>
                    </div>
                </td>
                <td>
                    <span class="badge ${event.Available > 0 ? 'bg-success' : 'bg-danger'}">
                        ${event.Available > 0 ? 'Активно' : 'Заполнено'}
                    </span>
                </td>
                <td>${this.formatDuration(event.BookingTTL)}</td>
                <td>
                    <span class="badge ${event.RequiresPayment ? 'bg-warning' : 'bg-info'}">
                        ${event.RequiresPayment ? 'Требуется' : 'Не требуется'}
                    </span>
                </td>
                <td>
                    <button class="btn btn-sm btn-outline-primary me-2"
                            onclick="app.viewEventDetails('${event.ID}')">
                        <i class="bi bi-eye"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger"
                            onclick="app.showCancelEventModal('${event.ID}', '${this.escapeHtml(event.Name)}')">
                        <i class="bi bi-trash"></i>
                    </button>
                </td>
            </tr>
        `).join('');
    }
    renderEventList(container) {
        if (this.events.length === 0) {
            container.innerHTML = '<div class="col-12"><p class="text-muted">Нет мероприятий</p></div>';
            return;
        }
        container.innerHTML = this.events.map(event => `
            <div class="col-md-6 mb-4">
                <div class="card">
                    <div class="card-header bg-primary text-white">
                        ${this.escapeHtml(event.Name)}
                    </div>
                    <div class="card-body">
                        <p><strong>Дата:</strong> ${this.formatDate(event.Date)}</p>
                        <p><strong>Доступно мест:</strong> ${event.Available} из ${event.TotalSeats}</p>
                        <p><strong>ID:</strong> <code>${event.ID}</code></p>
                        <button class="btn btn-sm btn-outline-secondary" onclick="app.copyToClipboard('${event.ID}')">
                            <i class="bi bi-clipboard"></i> Копировать ID
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }
    renderBookings() {
        const tableBody = document.getElementById('bookings-table-body');
        if (!tableBody) return;
        tableBody.innerHTML = this.bookings.map(booking => `
            <tr>
                <td><small>${booking.id.substring(0, 8)}...</small></td>
                <td>${this.escapeHtml(booking.event_name)}</td>
                <td>${booking.user_email}</td>
                <td>
                    <span class="badge ${this.getBookingStatusClass(booking.status)}">
                        ${this.getBookingStatusText(booking.status)}
                    </span>
                </td>
                <td>${this.formatTimeAgo(booking.created_at)}</td>
                <td>
                    ${booking.expires_at ? this.formatTimeRemaining(booking.expires_at) : '—'}
                </td>
                <td>
                    ${booking.status === 'pending' ? `
                        <button class="btn btn-sm btn-success me-1" onclick="app.confirmBooking('${booking.id}')">
                            <i class="bi bi-check"></i>
                        </button>
                    ` : ''}
                    <button class="btn btn-sm btn-danger" onclick="app.cancelBooking('${booking.id}')">
                        <i class="bi bi-x"></i>
                    </button>
                </td>
            </tr>
        `).join('');
    }
    updateStatistics() {
        // Обновление статистики на дашборде
        const eventsCount = this.events.length;
        const availableSeats = this.events.reduce((sum, event) => sum + event.Available, 0);
        const totalSeats = this.events.reduce((sum, event) => sum + event.TotalSeats, 0);
        const occupancyRate = totalSeats > 0 ? Math.round((totalSeats - availableSeats) / totalSeats * 100) : 0;
        // Обновление UI элементов
        const eventsCountEl = document.getElementById('events-count');
        const availableSeatsEl = document.getElementById('available-seats');
        const bookingProgressEl = document.getElementById('booking-progress');
        const progressRing = document.querySelector('.progress-ring-fill');
        if (eventsCountEl) eventsCountEl.textContent = eventsCount;
        if (availableSeatsEl) availableSeatsEl.textContent = availableSeats;
        if (bookingProgressEl) bookingProgressEl.textContent = `${occupancyRate}%`;
        // Анимация прогресс-кольца
        if (progressRing) {
            const radius = 45;
            const circumference = 2 * Math.PI * radius;
            const offset = circumference - (occupancyRate / 100) * circumference;
            progressRing.style.strokeDasharray = `${circumference} ${circumference}`;
            progressRing.style.strokeDashoffset = offset;
        }
        // Статистика для админ-панели
        const totalEventsEl = document.getElementById('total-events');
        const activeBookingsEl = document.getElementById('active-bookings');
        const pendingConfirmationsEl = document.getElementById('pending-confirmations');
        const expiredBookingsEl = document.getElementById('expired-bookings');
        if (totalEventsEl) totalEventsEl.textContent = eventsCount;
        if (activeBookingsEl) activeBookingsEl.textContent = totalSeats - availableSeats;
        if (pendingConfirmationsEl) {
            const pending = this.bookings.filter(b => b.status === 'pending').length;
            pendingConfirmationsEl.textContent = pending;
        }
        if (expiredBookingsEl) {
            const expired = this.bookings.filter(b =>
                b.status === 'pending' && new Date(b.expires_at) < new Date()
            ).length;
            expiredBookingsEl.textContent = expired;
        }
        // Обновление времени последнего обновления
        const lastUpdateEl = document.getElementById('last-update');
        if (lastUpdateEl) {
            lastUpdateEl.textContent = `Обновлено: ${new Date().toLocaleTimeString('ru-RU')}`;
        }
    }
    async handleRegister(e) {
        e.preventDefault();
       
        const formData = {
            email: document.getElementById('email').value,
            telegram: document.getElementById('telegram').value,
            role: document.querySelector('input[name="role"]:checked')?.value || 'user'
        };
       
        try {
            const response = await fetch(`${this.baseUrl}/users`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(formData)
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            const user = await response.json();
            this.currentUser = user;
           
            // Сохраняем в localStorage
            localStorage.setItem('eventbooker_user_id', user.ID);
            localStorage.setItem('eventbooker_user_email', user.Email);
           
            this.showToast(
                `Регистрация успешна! Ваш ID: ${user.ID}`,
                'success',
                'Регистрация'
            );
           
            this.updateUserUI();
           
            // Закрыть модальное окно и очистить форму
            const modal = bootstrap.Modal.getInstance(document.getElementById('registerModal'));
            if (modal) modal.hide();
            e.target.reset();
           
        } catch (error) {
            console.error('Registration error:', error);
            this.showToast(error.message || 'Ошибка регистрации', 'danger', 'Ошибка');
        }
    }
    async handleCreateEvent(e) {
        e.preventDefault();
       
        const formData = {
            name: document.getElementById('event-name').value,
            date: new Date(document.getElementById('event-date').value).toISOString(),
            total_seats: parseInt(document.getElementById('total-seats').value),
            booking_ttl: document.getElementById('booking-ttl').value,
            requires_payment: document.querySelector('input[name="requires-payment"]:checked').value === 'true'
        };
       
        try {
            const response = await fetch(`${this.baseUrl}/events`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(formData)
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            const event = await response.json();
           
            this.showToast(
                `Мероприятие "${event.Name}" создано!`,
                'success',
                'Создание мероприятия'
            );
           
            // Закрыть модальное окно и очистить форму
            const modal = bootstrap.Modal.getInstance(document.getElementById('createEventModal'));
            if (modal) modal.hide();
            e.target.reset();
           
            // Обновить список мероприятий
            setTimeout(() => this.loadEvents(), 1000);
           
        } catch (error) {
            console.error('Error creating event:', error);
            this.showToast(error.message || 'Ошибка создания мероприятия', 'danger', 'Ошибка');
        }
    }
    async handleBook(e) {
        e.preventDefault();
       
        const eventId = document.getElementById('book-event-id').value;
        const userId = document.getElementById('book-user-id').value;
       
        if (!eventId || !userId) {
            this.showToast('Заполните все поля', 'warning', 'Внимание');
            return;
        }
       
        const bookingData = { user_id: userId };
       
        try {
            const response = await fetch(`${this.baseUrl}/events/${eventId}/book`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(bookingData)
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            const booking = await response.json();
           
            this.showToast(
                `Бронь создана! ID брони: ${booking.ID}`,
                'success',
                'Бронирование'
            );
           
            // Закрыть модальное окно и очистить форму
            const modal = bootstrap.Modal.getInstance(document.getElementById('bookModal'));
            if (modal) modal.hide();
            e.target.reset();
           
            // Обновить список мероприятий
            setTimeout(() => this.loadEvents(), 1000);
           
        } catch (error) {
            console.error('Booking error:', error);
            this.showToast(error.message || 'Ошибка бронирования', 'danger', 'Ошибка');
        }
    }
    async handleConfirm(e) {
        e.preventDefault();
       
        const bookingId = document.getElementById('confirm-booking-id').value;
       
        if (!bookingId) {
            this.showToast('Введите ID брони', 'warning', 'Внимание');
            return;
        }
       
        try {
            const response = await fetch(`${this.baseUrl}/bookings/${bookingId}/confirm`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'}
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            this.showToast('Бронь успешно подтверждена!', 'success', 'Подтверждение');
           
            // Закрыть модальное окно и очистить форму
            const modal = bootstrap.Modal.getInstance(document.getElementById('confirmModal'));
            if (modal) modal.hide();
            e.target.reset();
           
            // Обновить список мероприятий
            setTimeout(() => this.loadEvents(), 1000);
           
        } catch (error) {
            console.error('Confirmation error:', error);
            this.showToast(error.message || 'Ошибка подтверждения', 'danger', 'Ошибка');
        }
    }
    async handleCancelEvent(e) {
        e.preventDefault();
       
        const eventId = document.getElementById('cancel-event-id').value;
        const reason = document.getElementById('cancel-reason').value;
       
        if (!eventId) {
            this.showToast('ID мероприятия не указан', 'warning', 'Внимание');
            return;
        }
       
        try {
            const response = await fetch(`${this.baseUrl}/events/${eventId}`, {
                method: 'DELETE',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({ reason })
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            this.showToast('Мероприятие успешно отменено', 'success', 'Отмена мероприятия');
           
            // Закрыть модальное окно и очистить форму
            const modal = bootstrap.Modal.getInstance(document.getElementById('cancelEventModal'));
            if (modal) modal.hide();
            e.target.reset();
           
            // Обновить список мероприятий
            setTimeout(() => this.loadEvents(), 1000);
           
        } catch (error) {
            console.error('Cancel event error:', error);
            this.showToast(error.message || 'Ошибка отмены мероприятия', 'danger', 'Ошибка');
        }
    }
    async confirmBooking(bookingId) {
        try {
            const response = await fetch(`${this.baseUrl}/bookings/${bookingId}/confirm`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'}
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            this.showToast('Бронь подтверждена', 'success', 'Подтверждение');
            setTimeout(() => this.loadBookings(), 1000);
           
        } catch (error) {
            console.error('Confirm booking error:', error);
            this.showToast(error.message || 'Ошибка подтверждения', 'danger', 'Ошибка');
        }
    }
    async cancelBooking(bookingId) {
        if (!confirm('Вы уверены, что хотите отменить эту бронь?')) return;
       
        try {
            const response = await fetch(`${this.baseUrl}/bookings/${bookingId}`, {
                method: 'DELETE',
                headers: {'Content-Type': 'application/json'}
            });
           
            if (!response.ok) {
                const error = await response.text();
                throw new Error(error);
            }
           
            this.showToast('Бронь отменена', 'success', 'Отмена брони');
            setTimeout(() => this.loadBookings(), 1000);
           
        } catch (error) {
            console.error('Cancel booking error:', error);
            this.showToast(error.message || 'Ошибка отмены брони', 'danger', 'Ошибка');
        }
    }
    // Вспомогательные методы
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
    formatDate(dateStr) {
        const date = new Date(dateStr);
        return date.toLocaleString('ru-RU', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }
    formatDuration(durationStr) {
        durationStr = durationStr.replace('0s', '');
        const match = durationStr.match(/(\d+)([mhd])/);
        if (!match) return durationStr;
       
        const value = parseInt(match[1]);
        const unit = match[2];
       
        switch(unit) {
            case 'm': return `${value} мин`;
            case 'h': return `${value} ч`;
            case 'd': return `${value} д`;
            default: return durationStr;
        }
    }
    formatTimeAgo(dateStr) {
        const date = new Date(dateStr);
        const now = new Date();
        const diff = Math.floor((now - date) / 1000);
       
        if (diff < 60) return 'только что';
        if (diff < 3600) return `${Math.floor(diff / 60)} мин назад`;
        if (diff < 86400) return `${Math.floor(diff / 3600)} ч назад`;
        return `${Math.floor(diff / 86400)} д назад`;
    }
    formatTimeRemaining(dateStr) {
        const date = new Date(dateStr);
        const now = new Date();
        const diff = Math.floor((date - now) / 1000);
       
        if (diff <= 0) return 'Истекло';
        if (diff < 60) return `${diff} сек`;
        if (diff < 3600) return `${Math.floor(diff / 60)} мин`;
        if (diff < 86400) return `${Math.floor(diff / 3600)} ч`;
        return `${Math.floor(diff / 86400)} д`;
    }
    getProgressColor(event) {
        const percent = (event.TotalSeats - event.Available) / event.TotalSeats * 100;
        if (percent < 50) return 'bg-success';
        if (percent < 80) return 'bg-warning';
        return 'bg-danger';
    }
    getOccupancyPercent(event) {
        return Math.round((event.TotalSeats - event.Available) / event.TotalSeats * 100);
    }
    getBookingStatusClass(status) {
        switch(status) {
            case 'confirmed': return 'bg-success';
            case 'pending': return 'bg-warning';
            case 'cancelled': return 'bg-danger';
            default: return 'bg-secondary';
        }
    }
    getBookingStatusText(status) {
        switch(status) {
            case 'confirmed': return 'Подтверждена';
            case 'pending': return 'Ожидает';
            case 'cancelled': return 'Отменена';
            default: return status;
        }
    }
    showToast(message, type = 'info', title = null) {
        const container = document.getElementById('alert-container') || this.createToastContainer();
       
        const toastId = 'toast-' + Date.now();
        const toast = document.createElement('div');
        toast.id = toastId;
        toast.className = `alert alert-${type} alert-dismissible fade show shadow`;
        toast.style.cssText = 'min-width: 300px; margin-bottom: 10px;';
        toast.innerHTML = `
            ${title ? `<strong>${title}</strong><br>` : ''}
            ${message}
            <button type="button" class="btn-close" onclick="this.parentElement.remove()"></button>
        `;
       
        container.appendChild(toast);
       
        // Автоудаление через 5 секунд
        setTimeout(() => {
            const element = document.getElementById(toastId);
            if (element) element.remove();
        }, 5000);
    }
    createToastContainer() {
        const container = document.createElement('div');
        container.id = 'alert-container';
        container.className = 'alert-toast';
        document.body.appendChild(container);
        return container;
    }
    showLoader() {
        const loader = document.getElementById('loader') || this.createLoader();
        loader.style.display = 'block';
    }
    hideLoader() {
        const loader = document.getElementById('loader');
        if (loader) loader.style.display = 'none';
    }
    createLoader() {
        const loader = document.createElement('div');
        loader.id = 'loader';
        loader.innerHTML = `
            <div style="position: fixed; top: 0; left: 0; width: 100%; height: 100%;
                       background: rgba(255,255,255,0.8); z-index: 9999; display: flex;
                       align-items: center; justify-content: center;">
                <div class="spinner-border text-primary" style="width: 3rem; height: 3rem;">
                    <span class="visually-hidden">Загрузка...</span>
                </div>
            </div>
        `;
        document.body.appendChild(loader);
        return loader;
    }
    filterEvents() {
        const searchTerm = document.getElementById('searchInput')?.value.toLowerCase() || '';
        const filteredEvents = this.events.filter(event =>
            event.Name.toLowerCase().includes(searchTerm) ||
            event.ID.toLowerCase().includes(searchTerm)
        );
       
        const tableBody = document.getElementById('events-table-body');
        if (tableBody) {
            tableBody.innerHTML = filteredEvents.map(event => `
                <tr>
                    <td>
                        <strong>${this.escapeHtml(event.Name)}</strong><br>
                        <small class="text-muted">ID: ${event.ID}</small>
                    </td>
                    <td>${this.formatDate(event.Date)}</td>
                    <td>
                        <div class="d-flex align-items-center">
                            <div class="progress flex-grow-1 me-2" style="height: 6px;">
                                <div class="progress-bar ${this.getProgressColor(event)}"
                                     style="width: ${this.getOccupancyPercent(event)}%">
                                </div>
                            </div>
                            <small>${event.Available}/${event.TotalSeats}</small>
                        </div>
                    </td>
                    <td>
                        <span class="badge ${event.Available > 0 ? 'bg-success' : 'bg-danger'}">
                            ${event.Available > 0 ? 'Активно' : 'Заполнено'}
                        </span>
                    </td>
                    <td>${this.formatDuration(event.BookingTTL)}</td>
                    <td>
                        <span class="badge ${event.RequiresPayment ? 'bg-warning' : 'bg-info'}">
                            ${event.RequiresPayment ? 'Требуется' : 'Не требуется'}
                        </span>
                    </td>
                    <td>
                        <button class="btn btn-sm btn-outline-primary me-2"
                                onclick="app.viewEventDetails('${event.ID}')">
                            <i class="bi bi-eye"></i>
                        </button>
                        <button class="btn btn-sm btn-outline-danger"
                                onclick="app.showCancelEventModal('${event.ID}', '${this.escapeHtml(event.Name)}')">
                            <i class="bi bi-trash"></i>
                        </button>
                    </td>
                </tr>
            `).join('');
        }
    }
    filterBookings(filter) {
        const tableBody = document.getElementById('bookings-table-body');
        if (!tableBody) return;
       
        let filtered = this.bookings;
        if (filter !== 'all') {
            filtered = this.bookings.filter(b => b.status === filter);
        }
       
        tableBody.innerHTML = filtered.map(booking => `
            <tr>
                <td><small>${booking.id.substring(0, 8)}...</small></td>
                <td>${this.escapeHtml(booking.event_name)}</td>
                <td>${booking.user_email}</td>
                <td>
                    <span class="badge ${this.getBookingStatusClass(booking.status)}">
                        ${this.getBookingStatusText(booking.status)}
                    </span>
                </td>
                <td>${this.formatTimeAgo(booking.created_at)}</td>
                <td>
                    ${booking.expires_at ? this.formatTimeRemaining(booking.expires_at) : '—'}
                </td>
                <td>
                    ${booking.status === 'pending' ? `
                        <button class="btn btn-sm btn-success me-1" onclick="app.confirmBooking('${booking.id}')">
                            <i class="bi bi-check"></i>
                        </button>
                    ` : ''}
                    <button class="btn btn-sm btn-danger" onclick="app.cancelBooking('${booking.id}')">
                        <i class="bi bi-x"></i>
                    </button>
                </td>
            </tr>
        `).join('');
    }
    showCancelEventModal(eventId, eventName) {
        document.getElementById('cancel-event-id').value = eventId;
        document.getElementById('cancel-event-name').textContent = eventName;
       
        const modal = new bootstrap.Modal(document.getElementById('cancelEventModal'));
        modal.show();
    }
    quickBook(eventId) {
        document.getElementById('book-event-id').value = eventId;
        const modal = new bootstrap.Modal(document.getElementById('bookModal'));
        modal.show();
    }
    copyToClipboard(text) {
        navigator.clipboard.writeText(text).then(() => {
            this.showToast('Скопировано в буфер обмена', 'success');
        });
    }
    startAutoRefresh() {
        // Обновление каждые 30 секунд для админа, 60 для пользователя
        const interval = this.currentPage === 'admin' ? 30000 : 60000;
        setInterval(() => {
            this.loadEvents();
            if (this.currentPage === 'admin') {
                this.loadBookings();
            }
        }, interval);
    }
    updateUI() {
        // Показать/скрыть элементы в зависимости от страницы
        if (this.currentPage === 'admin') {
            document.body.classList.add('admin-panel');
            this.loadBookings();
        } else {
            document.body.classList.add('user-panel');
        }
    }
    async viewEventDetails(eventId) {
        try {
            const response = await fetch(`${this.baseUrl}/events/${eventId}`);
            if (!response.ok) throw new Error('Ошибка загрузки деталей');
            const event = await response.json();
            // Render in modal
            document.getElementById('event-details-title').textContent = event.Name;
            document.getElementById('event-details-content').innerHTML = `
                <p><strong>ID:</strong> ${event.ID}</p>
                <p><strong>Дата:</strong> ${this.formatDate(event.Date)}</p>
                <p><strong>Места:</strong> ${event.Available} / ${event.TotalSeats}</p>
                <p><strong>TTL:</strong> ${this.formatDuration(event.BookingTTL)}</p>
                <p><strong>Статус:</strong> ${event.Status}</p>
            `;
            const modal = new bootstrap.Modal(document.getElementById('eventDetailsModal'));
            modal.show();
        } catch (error) {
            this.showToast('Ошибка загрузки деталей', 'danger');
        }
    }
}
// Инициализация приложения
const app = new EventBooker();
// Глобальные функции для использования в onclick
window.app = app;
window.copyToClipboard = (text) => app.copyToClipboard(text);
window.quickBook = (eventId) => app.quickBook(eventId);
window.confirmBooking = (bookingId) => app.confirmBooking(bookingId);
window.cancelBooking = (bookingId) => app.cancelBooking(bookingId);
window.viewEventDetails = (eventId) => app.viewEventDetails(eventId);
window.showCancelEventModal = (eventId, eventName) => app.showCancelEventModal(eventId, eventName);