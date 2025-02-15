document.addEventListener("DOMContentLoaded", function () {
    console.log("📢 Страница загружена!");
    checkAuthStatus();
    loadFavorites();
    
});


  function showSection(sectionId) {
    var section = document.getElementById(sectionId);
    if (section) {
      
      var headerOffset = 120; 
      var elementPosition = section.getBoundingClientRect().top;
      var offsetPosition = elementPosition + window.pageYOffset - headerOffset;
      window.scrollTo({
        top: offsetPosition,
        behavior: "smooth"
      });
    }
  }

// 🌟 Функция поиска лекарства
async function searchMedicine() {
    let query = document.getElementById('search').value.trim();
    let resultsList = document.getElementById('results');
    resultsList.innerHTML = '';

    if (!query) {
        alert("Введите название лекарства");
        return;
    }

    try {
        let response = await fetch(`/medicines/search?q=${encodeURIComponent(query)}`);

        if (!response.ok) {
            throw new Error("Ошибка сервера. Попробуйте позже.");
        }

        let data = await response.json();

        if (!Array.isArray(data) || data.length === 0) {
            alert("Лекарство не найдено!");
            return;
        }

        data.forEach(med => {
            if (!med.id) {
                console.error("❌ Ошибка: у лекарства нет ID", med);
                return;
            }

            let li = document.createElement('li');
            li.innerHTML = `
                <img src="${med.image_url || 'https://via.placeholder.com/100'}" 
                     alt="${med.name}" 
                     style="width: 100px; height: 100px;" 
                     onerror="this.onerror=null; this.src='https://via.placeholder.com/100';">
                <p><b>${med.name}</b> - ${med.description} (Категория: ${med.category}, Цена: $${med.price})</p>
                <button onclick="addToFavorites(${med.id})">⭐ В избранное</button>
            `;
            resultsList.appendChild(li);
        });
    } catch (error) {
        console.error("❌ Ошибка при поиске лекарства:", error);
        alert("Ошибка загрузки данных. Попробуйте снова!");
    }
}
function filterByCategory() {
    var category = document.getElementById('medicine-category').value;
    if (!category) {
      alert("Пожалуйста, выберите категорию");
      return;
    }
    
    fetch('/medicines/category?category=' + encodeURIComponent(category))
      .then(response => response.json())
      .then(data => {
        const results = document.getElementById('results');
        results.innerHTML = "";
        
        if (!Array.isArray(data) || data.length === 0) {
          results.innerHTML = "<li>Лекарства не найдены</li>";
          return;
        }
        
        data.forEach(med => {
          let li = document.createElement('li');
          li.innerHTML = `
            <img src="${med.image_url}" alt="${med.name}" style="width: 100px; height: 100px;">
            <strong>${med.name}</strong> - ${med.description}
          `;
          results.appendChild(li);
        });
      })
      .catch(error => console.error('Ошибка при получении данных:', error));
  }
  
  window.filterByCategory = filterByCategory;
  

// 🌟 Функция переключения модального окна
function toggleModal(modalId) {
    let modal = document.getElementById(modalId);
    if (modal) {
        modal.style.display = modal.style.display === "block" ? "none" : "block";
    }
}

// 🌟 Регистрация пользователя
async function registerUser() {
 
    let firstName = document.getElementById("first-name").value.trim();
    let lastName = document.getElementById("last-name").value.trim();
    let email = document.getElementById("email").value.trim(); 
    let phone = document.getElementById("phone").value.trim();
    let city = document.getElementById("city").value.trim();
    let password = document.getElementById("password").value.trim();
    let confirmPassword = document.getElementById("confirm-password").value.trim();
    let message = document.getElementById("message");

    
    if (!firstName || !lastName || !email || !phone || !city || !password || !confirmPassword) {
        message.textContent = "Все поля обязательны!";
        message.style.color = "red";
        return;
    }

  
    if (!email.includes("@")) {
        message.textContent = "Email должен содержать символ @!";
        message.style.color = "red";
        return;
    }

  
    if (password !== confirmPassword) {
        message.textContent = "Пароли не совпадают!";
        message.style.color = "red";
        return;
    }

    try {
        let response = await fetch('/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                first_name: firstName,
                last_name: lastName,
                email: email,      
                phone: phone,
                city: city,
                password: password
            })
        });

        let data = await response.json();

        if (response.ok) {
            message.textContent = "Регистрация успешна!";
            message.style.color = "green";
            toggleModal('register-modal');
        } else {
            message.textContent = data.error || "Ошибка регистрации!";
            message.style.color = "red";
        }
    } catch (error) {
        console.error("Ошибка при регистрации:", error);
        message.textContent = "Ошибка соединения!";
        message.style.color = "red";
    }
}


function renderClinicCards(clinics) {
    const container = document.getElementById("clinic-cards");
    container.innerHTML = "";
  
    if (!Array.isArray(clinics) || clinics.length === 0) {
      container.innerHTML = "<p>Клиники не найдены</p>";
      return;
    }
    
    clinics.forEach(clinic => {
      const card = document.createElement("div");
      card.className = "clinic-card";
      card.innerHTML = `
        <img src="${clinic.image_url}" alt="${clinic.name}">
        <div class="clinic-info">
          <h3>${clinic.name}</h3>
          <p class="address">${clinic.address}</p>
          <p class="website"><a href="${clinic.url}" target="_blank">${clinic.url}</a></p>
          <p class="description">${clinic.description}</p>
        </div>
      `;
      container.appendChild(card);
    });
  }
  
  // Функция фильтрации клиник 
  function filterClinicsByCity() {
    const city = document.getElementById('clinic-city').value;
    if (!city) {
      alert("Пожалуйста, выберите город");
      return;
    }
    
    // Выполняем запрос к серверу
    fetch('/clinics?city=' + encodeURIComponent(city))
      .then(response => {
       
        if (!response.ok) {
          throw new Error("Ошибка сервера: " + response.status);
        }
        return response.json();
      })
      .then(data => {
    
        renderClinicCards(data);
      })
      .catch(error => console.error("Ошибка при получении клиник:", error));
  }
  
  window.filterClinicsByCity = filterClinicsByCity;
  

  document.addEventListener("DOMContentLoaded", function () {

    const city = document.getElementById('clinic-city').value;
    if (city) {
      filterClinicsByCity();
    } else {
      // Если город не выбран, можно отобразить сообщение или загрузить все клиники,
      // если на сервере реализована логика для пустого значения.
      // Например, можно сделать:
      fetch('/clinics')
        .then(response => response.json())
        .then(data => renderClinicCards(data))
        .catch(error => console.error("Ошибка при получении всех клиник:", error));
    }
  });
  
  
  
  

// 🌟 Вход в аккаунт
async function loginUser() {
    let loginInput = document.getElementById("login-input")?.value.trim();
    let password = document.getElementById("login-password")?.value.trim();
    let message = document.getElementById("login-message");

    if (!loginInput || !password) {
        message.textContent = "Введите логин и пароль!";
        message.style.color = "red";
        return;
    }

    try {
        let response = await fetch('/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ login: loginInput, password })
        });

        let data = await response.json();
        console.log("Ответ сервера (login):", data);

        if (response.ok && data.user_id) {
            localStorage.setItem("user_id", data.user_id);
            console.log("User id сохранён:", localStorage.getItem("user_id"));
            message.textContent = "Вход выполнен!";
            message.style.color = "green";
            toggleModal('login-modal');
            checkAuthStatus();
            loadFavorites();
        } else {
            message.textContent = data.error || "Неверные учетные данные!";
            message.style.color = "red";
        }
    } catch (error) {
        console.error("Ошибка входа:", error);
        message.textContent = "Ошибка сервера. Попробуйте снова!";
        message.style.color = "red";
    }
}


// 🌟 Функция добавления в избранное для авторизованных
async function addToFavorites(medicineId) {
    let userId = localStorage.getItem("user_id");

    if (!userId) {
        alert("Сначала войдите в систему!");
        return;
    }

    try {
        let response = await fetch('/favorites', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: parseInt(userId), medicine_id: parseInt(medicineId) })
        });

        let data = await response.json();
        alert(data.message);
        loadFavorites();
    } catch (error) {
        console.error("❌ Ошибка добавления в избранное:", error);
        alert("Ошибка сервера.");
    }
}




//  Функция выхода
function logoutUser() {
    localStorage.removeItem("user_id");
    alert("Вы вышли из аккаунта!");
    checkAuthStatus();
    location.reload();
}



async function loadFavorites() {
    let userId = localStorage.getItem("user_id");
    let favoritesList = document.getElementById("favorites-list");

    if (!userId || !favoritesList) {
        favoritesList.innerHTML = "<p>Войдите, чтобы увидеть избранное.</p>";
        return;
    }

    try {
        let response = await fetch(`/favorites?user_id=${userId}`);
        let data = await response.json();
        favoritesList.innerHTML = '';

        if (!Array.isArray(data) || data.length === 0) {
            favoritesList.innerHTML = "<p>Избранное пусто.</p>";
            return;
        }

        data.forEach(fav => {
            if (!fav.medicine) return; 

            let li = document.createElement("li");
            li.innerHTML = `
                <img src="${fav.medicine.image_url || 'https://via.placeholder.com/100'}" 
                     alt="${fav.medicine.name}" 
                     style="width: 80px; height: 100px;" 
                     onerror="this.onerror=null; this.src='https://via.placeholder.com/100';">
                <p><b>${fav.medicine.name}</b> - ${fav.medicine.description} 
                (Категория: ${fav.medicine.category}, Цена: KZT ${fav.medicine.price})</p>
                <button onclick="removeFromFavorites(${fav.id})">Удалить</button>
            `;
            favoritesList.appendChild(li);
        });
    } catch (error) {
        console.error(" Ошибка загрузки избранного:", error);
    }
}

// 🌟 Удаление лекарства из избранного
async function removeFromFavorites(favoriteId) {
    try {
        let response = await fetch(`/favorites/${favoriteId}`, {
            method: 'DELETE'
        });

        let data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || "Ошибка при удалении!");
        }

        alert(data.message);
        loadFavorites(); 
    } catch (error) {
        console.error("Ошибка удаления из избранного:", error);
        alert("Не удалось удалить лекарство!");
    }
}

  
// Вызываем при загрузке страницы и после входа/выхода
function checkAuthStatus() {
    const userId = localStorage.getItem("user_id");
    const authButtons = document.querySelector(".auth-buttons");

    if (userId) {
        authButtons.innerHTML = `
            <button class="profile-btn" onclick="showProfile()">Профиль</button>
            <button class="logout-btn" onclick="logoutUser()">Выйти</button>
        `;
        loadUserProfile(); 
    } else {
        authButtons.innerHTML = `
            <button class="login-btn" onclick="toggleModal('login-modal')">Войти</button>
            <button class="register-btn" onclick="toggleModal('register-modal')">Регистрация</button>
        `;
        document.getElementById("user-info").textContent = "Вы не авторизованы";
    }
}



function showProfile() {
    
    toggleModal('profile-modal');
    // Загружаем данные профиля для модального окна
    loadUserProfileModal();
}
async function loadUserProfileModal() {
    const userId = localStorage.getItem("user_id");
    console.log("loadUserProfileModal: userId =", userId); 

    if (!userId) {
        document.getElementById("profile-details").textContent = "Вы не авторизованы";
        return;
    }

    try {
        console.log("Отправляем запрос: `/profile?user_id=" + userId + "`");
        const response = await fetch(`/profile?user_id=${userId}`);
        console.log("Response status:", response.status);

        if (!response.ok) {
            throw new Error("Ошибка получения профиля");
        }

        const data = await response.json();
        console.log("Полученные данные профиля:", data);

        document.getElementById("profile-details").innerHTML = `
            <p><strong>Имя:</strong> ${data.first_name} ${data.last_name}</p>
            <p><strong>Город:</strong> ${data.city}</p>
            ${data.email ? `<p><strong>Email:</strong> ${data.email}</p>` : ""}
            ${data.phone ? `<p><strong>Телефон:</strong> ${data.phone}</p>` : ""}
        `;
    } catch (error) {
        console.error("Ошибка загрузки профиля (modal):", error);
        document.getElementById("profile-details").textContent = "Не удалось загрузить профиль.";
    }
}

async function loadUserProfile() {
    const userId = localStorage.getItem("user_id");
    if (!userId) {
        document.getElementById("user-info").textContent = "Вы не авторизованы";
        return;
    }

    try {
        const response = await fetch(`/profile?user_id=${userId}`);
        if (!response.ok) {
            throw new Error("Ошибка получения профиля");
        }

        const data = await response.json();
        console.log("Данные профиля:", data);

    
        document.getElementById("user-info").innerHTML = `<strong>Добро пожаловать, ${data.first_name} ${data.last_name}!</strong>`;
    } catch (error) {
        console.error("Ошибка загрузки профиля:", error);
        document.getElementById("user-info").textContent = "Ошибка загрузки профиля";
    }
}
 
// Функция загрузки данных профиля для модального окна
async function loadUserProfileModal() {
    const userId = localStorage.getItem("user_id");
    if (!userId) {
        document.getElementById("profile-details").innerHTML = "Вы не авторизованы";
        return;
    }

    try {
        const response = await fetch(`/profile?user_id=${userId}`);
        if (!response.ok) {
            throw new Error("Ошибка получения профиля");
        }

        const data = await response.json();

        document.getElementById("profile-first-name").textContent = data.first_name || "Не указано";
        document.getElementById("profile-last-name").textContent = data.last_name || "";
        document.getElementById("profile-email").textContent = data.email || "Не указано";
        document.getElementById("profile-phone").textContent = data.phone || "Не указано";
        document.getElementById("profile-city").textContent = data.city || "Не указано";
    } catch (error) {
        console.error("Ошибка загрузки профиля:", error);
        document.getElementById("profile-details").innerHTML = "Ошибка загрузки профиля";
    }
}


function enableProfileEditing() {
    document.getElementById("edit-profile-form").style.display = "block";

    // Заполняем поля текущими значениями
    const profileText = document.getElementById("profile-details").innerText;
    const nameMatch = profileText.match(/Имя:\s(.*)/);
    const emailMatch = profileText.match(/Email:\s(.*)/);
    const phoneMatch = profileText.match(/Телефон:\s(.*)/);

    document.getElementById("edit-first-name").value = nameMatch ? nameMatch[1].split(" ")[0] : "";
    document.getElementById("edit-last-name").value = nameMatch ? nameMatch[1].split(" ")[1] : "";
    document.getElementById("edit-email").value = emailMatch ? emailMatch[1] : "";
    document.getElementById("edit-phone").value = phoneMatch ? phoneMatch[1] : "";
}

async function saveProfileChanges() {
    const userId = localStorage.getItem("user_id");
    if (!userId) return alert("Ошибка: Вы не авторизованы!");

    const firstName = document.getElementById("edit-first-name").value.trim();
    const lastName = document.getElementById("edit-last-name").value.trim();
    const email = document.getElementById("edit-email").value.trim();
    const phone = document.getElementById("edit-phone").value.trim();

    if (!firstName || !lastName || !email || !phone) {
        alert("Все поля обязательны!");
        return;
    }

    try {
        const response = await fetch(`/update-profile`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ user_id: userId, first_name: firstName, last_name: lastName, email, phone })
        });

        const data = await response.json();
        if (response.ok) {
            alert("Профиль обновлён!");
            loadUserProfile();
            toggleModal("profile-modal");
        } else {
            alert("Ошибка: " + data.error);
        }
    } catch (error) {
        console.error("Ошибка обновления профиля:", error);
        alert("Ошибка сервера!");
    }
}



async function saveProfileChanges() {
    const userId = parseInt(localStorage.getItem("user_id"), 10);
    if (!userId || isNaN(userId)) {
        alert("Ошибка: Вы не авторизованы!");
        return;
    }

    const firstName = document.getElementById("edit-first-name").value.trim();
    const lastName = document.getElementById("edit-last-name").value.trim();
    const email = document.getElementById("edit-email").value.trim();
    const phone = document.getElementById("edit-phone").value.trim();

    let updatedData = { user_id: userId };
    if (firstName) updatedData.first_name = firstName;
    if (lastName) updatedData.last_name = lastName;
    if (email) updatedData.email = email;
    if (phone) updatedData.phone = phone;

    if (Object.keys(updatedData).length === 1) {
        alert("Вы не изменили ни одного поля!");
        return;
    }

    console.log("Отправка запроса на сервер...");
    console.log("Тело запроса JSON:", JSON.stringify(updatedData, null, 2));

    try {
        const response = await fetch(`/update-profile`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(updatedData)
        });

        const data = await response.json();
        console.log("Ответ сервера:", data);

        if (response.ok) {
            alert("Профиль обновлён!");
            loadUserProfile(); 
            toggleModal("profile-modal");
        } else {
            alert("Ошибка: " + data.error);
        }
    } catch (error) {
        console.error("Ошибка обновления профиля:", error);
        alert("Ошибка сервера!");
    }
}


function deleteProfile() {
    if (confirm("Вы уверены, что хотите удалить профиль?")) {
        const userId = localStorage.getItem("user_id"); // Adjust based on how you store user ID
        if (!userId) {
            alert("Ошибка: Не удалось получить user_id.");
            return;
        }

        fetch('/delete-profile', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: parseInt(userId) })
        })
        .then(response => response.json())
        .then(data => {
            if (data.message) {
                alert("Профиль удален!");
                localStorage.clear(); // Clear user data from local storage
                window.location.href = "/"; // Redirect to homepage
            } else {
                alert("Ошибка: " + data.error);
            }
        })
        .catch(error => console.error('Ошибка:', error));
    }
}
  


const faqQuestions = document.querySelectorAll('.faq-question');
faqQuestions.forEach(question => {
  question.addEventListener('click', function() {
    this.classList.toggle('active');
    const answer = this.nextElementSibling;
    answer.classList.toggle('open');
  });
});















  

