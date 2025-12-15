function updateEvents(listId) {
    fetch('/api/events')
        .then(res => res.json())
        .then(data => {
            const list = document.getElementById(listId);
            list.innerHTML = '';
            data.forEach(e => {
                const p = document.createElement('p');
                p.innerText = `ID: ${e.ID}, Name: ${e.Name}, Available: ${e.Available}, Date: ${new Date(e.Date).toLocaleString()}`;
                list.appendChild(p);
            });
        })
        .catch(err => console.error(err));
}

function setupRegisterForm() {
    const form = document.getElementById('register-form');
    if (form) {
        form.addEventListener('submit', e => {
            e.preventDefault();
            const email = document.getElementById('email').value;
            const telegram = document.getElementById('telegram').value;
            fetch('/api/users', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({email, telegram, role: 'user'})
            }).then(res => res.json()).then(data => {
                alert('User registered: ' + data.ID);
            }).catch(err => alert(err));
        });
    }
}

function setupBookForm() {
    const form = document.getElementById('book-form');
    if (form) {
        form.addEventListener('submit', e => {
            e.preventDefault();
            const eventId = document.getElementById('book-event-id').value;
            const userId = document.getElementById('book-user-id').value;
            fetch(`/api/events/${eventId}/book`, {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({user_id: userId})
            }).then(res => res.json()).then(data => {
                alert('Booked: ' + data.ID);
                updateEvents('events-list');
            }).catch(err => alert(err));
        });
    }
}

function setupConfirmForm() {
    const form = document.getElementById('confirm-form');
    if (form) {
        form.addEventListener('submit', e => {
            e.preventDefault();
            const bookingId = document.getElementById('confirm-booking-id').value;
            fetch(`/api/bookings/${bookingId}/confirm`, {method: 'POST'}).then(res => {
                if (res.ok) {
                    alert('Confirmed');
                    updateEvents('events-list');
                } else {
                    alert('Error');
                }
            }).catch(err => alert(err));
        });
    }
}

function setupCreateEventForm() {
    const form = document.getElementById('create-event-form');
    if (form) {
        form.addEventListener('submit', e => {
            e.preventDefault();
            const name = document.getElementById('name').value;
            const date = new Date(document.getElementById('date').value);
            const totalSeats = parseInt(document.getElementById('total-seats').value);
            const bookingTTL = document.getElementById('booking-ttl').value;
            const requiresPayment = document.getElementById('requires-payment').checked;
            fetch('/api/events', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({name, date, total_seats: totalSeats, booking_ttl: bookingTTL, requires_payment: requiresPayment})
            }).then(res => res.json()).then(data => {
                alert('Event created: ' + data.ID);
                updateEvents('events-list');
            }).catch(err => alert(err));
        });
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const eventsList = document.getElementById('events-list');
    if (eventsList) {
        updateEvents('events-list');
        setInterval(() => updateEvents('events-list'), 10000); // Poll every 10s to track expirations
    }
    setupRegisterForm();
    setupBookForm();
    setupConfirmForm();
    setupCreateEventForm();
});