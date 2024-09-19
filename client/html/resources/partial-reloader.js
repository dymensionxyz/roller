const reloadPartialBalances = () => {
    fetch('/partial/balances')
        .then(response => response.text())
        .then(html => {
            document.getElementById('partial-balances').innerHTML = html;
        });
}