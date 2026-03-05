function toggleDropdown() {
    const dropdown = document.getElementById('userDropdown');
    dropdown.classList.toggle('show');
}

// Fecha o menu se clicar fora dele
window.onclick = function(event) {
    if (!event.target.closest('.user-menu-container')) {
        const dropdowns = document.getElementsByClassName("user-dropdown");
        for (let i = 0; i < dropdowns.length; i++) {
            const openDropdown = dropdowns[i];
            if (openDropdown.classList.contains('show')) {
                openDropdown.classList.remove('show');
            }
        }
    }
}