async function searchMedicine() {
    let query = document.getElementById('search').value.trim();
    let errorMessage = document.getElementById('error-message');
    let notFoundMessage = document.getElementById('not-found-message');
    let resultsList = document.getElementById('results');
    
    resultsList.innerHTML = '';
    errorMessage.style.display = 'none';
    notFoundMessage.style.display = 'none';

    if (!query) {
        errorMessage.style.display = 'block';
        return;
    }

    try {
        let response = await fetch(`/medicines/search?q=${query}`);
        if (!response.ok) {
            throw new Error("Server error, please try again later.");
        }
        
        let data = await response.json();

        if (!data || data.length === 0) {
            notFoundMessage.style.display = 'block';
            return;
        }

        data.forEach(med => {
            let li = document.createElement('li');
            li.textContent = `${med.name} - ${med.description} (категория: ${med.category}, примерная цена: $${med.price})`;
            resultsList.appendChild(li);
        });
    } catch (error) {
        console.error("enter the correct name the drug:", error);
        notFoundMessage.textContent = "Error to find drug. Please try again.";
        notFoundMessage.style.display = 'block';
    }
}
