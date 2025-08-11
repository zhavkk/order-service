document.getElementById('order-form').addEventListener('submit', async (event) => {
    event.preventDefault();
    const orderId = document.getElementById('order-id').value;
    const resultDiv = document.getElementById('order-result');
    resultDiv.innerHTML = 'Loading...';

    try {
        const response = await fetch(`http://localhost:8080/orders/${orderId}`);
        if (!response.ok) {
            throw new Error(`Error: ${response.status}`);
        }
        const data = await response.json();
        resultDiv.innerHTML = `<pre>${JSON.stringify(data, null, 2)}</pre>`;
    } catch (error) {
        resultDiv.innerHTML = `<p style="color: red;">${error.message}</p>`;
    }
});