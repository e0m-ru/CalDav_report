document.querySelectorAll('#sortableTable th.sortable').forEach(header => {
        header.addEventListener('click', () => {
            const table = header.closest('table');
            const columnIndex = header.cellIndex;
            const dataType = header.dataset.type || 'string';
            const isAsc = header.classList.contains('asc');
            const direction = isAsc ? 'desc' : 'asc';

            sortTable(table, columnIndex, dataType, direction);

            // Сбрасываем классы сортировки
            table.querySelectorAll('th').forEach(th => {
                th.classList.remove('asc', 'desc');
            });

            header.classList.add(direction);
        });
});

    function sortTable(table, columnIndex, dataType, direction) {
	const tbody = table.tBodies[0];
    const rows = Array.from(tbody.querySelectorAll('tr'));

	rows.sort((a, b) => {
		const aVal = a.cells[columnIndex].textContent;
    const bVal = b.cells[columnIndex].textContent;

    switch (dataType) {
			case 'number':
    return compareNumbers(aVal, bVal, direction);
    case 'date':
    return compareDates(aVal, bVal, direction);
    default:
    return compareStrings(aVal, bVal, direction);
		}
	});

    tbody.innerHTML = '';
	rows.forEach(row => tbody.appendChild(row));
}

    function compareNumbers(a, b, direction) {
	const numA = parseFloat(a);
    const numB = parseFloat(b);
    return direction === 'asc' ? numA - numB : numB - numA;
}

    function compareDates(a, b, direction) {
	const dateA = new Date(a);
    const dateB = new Date(b);
    return direction === 'asc' ? dateA - dateB : dateB - dateA;
}

    function compareStrings(a, b, direction) {
	return direction === 'asc'
    ? a.localeCompare(b, 'ru')
    : b.localeCompare(a, 'ru');
}