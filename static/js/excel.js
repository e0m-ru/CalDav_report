function exportToExcelXLSX() {
    // Создаем клон таблицы для модификации перед экспортом
    const table = document.getElementById("sortableTable");
    const clonedTable = table.cloneNode(true);

    // Обрабатываем все ячейки с чекбоксами
    const cellsWithCheckboxes = clonedTable.querySelectorAll('td:has(input[type="checkbox"])');

    cellsWithCheckboxes.forEach(td => {
        const checkbox = td.querySelector('input[type="checkbox"]');
        const label = td.querySelector('label');

        if (checkbox.checked) {
            // Если чекбокс отмечен, оставляем текст label
            td.textContent = label.textContent;
        } else {
            // Если чекбокс не отмечен, очищаем содержимое
            td.textContent = "";
        }
    });

    // Конвертируем модифицированную таблицу в Excel
    const workbook = XLSX.utils.table_to_book(clonedTable);
    XLSX.writeFile(workbook, "sortableTable.xlsx");
}
