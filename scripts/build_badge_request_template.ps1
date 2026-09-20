$ErrorActionPreference = 'Stop'

$outputPath = Join-Path (Split-Path -Parent $PSScriptRoot) 'Шаблон_заявки_на_аккредитацию_СПО_ВО.xlsx'
$spoSpecialties = @(
    'Анестезиология и реаниматология', 'Гистология', 'Лечебная физкультура',
    'Лечебное дело', 'Медицинский массаж', 'Медицинский массаж ОВЗ',
    'Операционное дело', 'Реабилитационное сестринское дело', 'Рентгенология',
    'Сестринское дело', 'Сестринское дело в педиатрии',
    'Скорая и неотложная помощь', 'Техподдержки', 'Приём документов',
    'Физиотерапия', 'Функциональная диагностика'
)
$roles = @(
    'Председатель подкомиссии', 'Заместитель председателя подкомиссии',
    'Секретарь подкомиссии', 'Член подкомиссии'
)

function Set-RequestSheet {
    param([object]$Sheet, [string]$Profile, [string]$SpecialtyListName)

    $lastRow = 204
    $Sheet.Cells.Font.Name = 'Calibri'
    $Sheet.Cells.Font.Size = 11
    $Sheet.Range('A1:C1').Merge()
    $Sheet.Range('A1').Value2 = "Заявка на аккредитацию — $Profile"
    $Sheet.Range('A1').Font.Bold = $true
    $Sheet.Range('A1').Font.Size = 16
    $Sheet.Range('A1').HorizontalAlignment = -4108
    $Sheet.Range('A1').Interior.Color = 4533685
    $Sheet.Range('A1').RowHeight = 28

    $Sheet.Range('A2:C2').Merge()
    $Sheet.Range('A2').Value2 = 'Одна строка — один член подкомиссии. Если у человека несколько ролей или специальностей, укажите его отдельными строками.'
    $Sheet.Range('A2').WrapText = $true
    $Sheet.Range('A2').Interior.Color = 15921906
    $Sheet.Range('A2').RowHeight = 30

    $Sheet.Range('A4').Value2 = 'ФИО полностью'
    $Sheet.Range('B4').Value2 = 'Роль в подкомиссии'
    $Sheet.Range('C4').Value2 = 'Специальность'
    $header = $Sheet.Range('A4:C4')
    $header.Font.Bold = $true
    $header.Font.Color = 16777215
    $header.Interior.Color = 4533685
    $header.HorizontalAlignment = -4108
    $header.VerticalAlignment = -4108
    $header.RowHeight = 25

    $dataRange = $Sheet.Range("A5:C$lastRow")
    $dataRange.Borders.LineStyle = 1
    $dataRange.Borders.Color = 14277081
    $dataRange.VerticalAlignment = -4108
    $dataRange.WrapText = $true
    $dataRange.RowHeight = 22
    $dataRange.Interior.Color = 16777215
    $Sheet.Columns.Item(1).ColumnWidth = 35
    $Sheet.Columns.Item(2).ColumnWidth = 36
    $Sheet.Columns.Item(3).ColumnWidth = 42

    $roleRange = $Sheet.Range("B5:B$lastRow")
    $roleRange.Validation.Delete()
    $roleRange.Validation.Add(3, 1, 1, '=Roles')
    $roleRange.Validation.IgnoreBlank = $true
    $roleRange.Validation.InCellDropdown = $true
    $roleRange.Validation.ErrorTitle = 'Выберите роль из списка'
    $roleRange.Validation.ErrorMessage = 'Свободный текст в поле «Роль» не допускается.'

    $specialtyRange = $Sheet.Range("C5:C$lastRow")
    $specialtyRange.Validation.Delete()
    $specialtyRange.Validation.Add(3, 1, 1, "=$SpecialtyListName")
    $specialtyRange.Validation.IgnoreBlank = $true
    $specialtyRange.Validation.InCellDropdown = $true
    $specialtyRange.Validation.ErrorTitle = 'Выберите специальность из списка'
    $specialtyRange.Validation.ErrorMessage = 'Если специальности нет, добавьте её на листе «Настройки».'

    $Sheet.Activate()
    $window = $Sheet.Parent.Windows.Item(1)
    $window.SplitRow = 4
    $window.FreezePanes = $true
}

function Set-SettingsSheet {
    param([object]$Sheet)

    $Sheet.Cells.Font.Name = 'Calibri'
    $Sheet.Cells.Font.Size = 11
    $Sheet.Range('A1:C1').Merge()
    $Sheet.Range('A1').Value2 = 'Настройки: справочники для выпадающих списков'
    $Sheet.Range('A1').Font.Bold = $true
    $Sheet.Range('A1').Font.Size = 14
    $Sheet.Range('A1').Interior.Color = 4533685
    $Sheet.Range('A1').HorizontalAlignment = -4108
    $Sheet.Range('A3').Value2 = 'Роли'
    $Sheet.Range('B3').Value2 = 'Специальности СПО'
    $Sheet.Range('C3').Value2 = 'Специальности ВО'
    $Sheet.Range('A3:C3').Font.Bold = $true
    $Sheet.Range('A3:C3').Font.Color = 16777215
    $Sheet.Range('A3:C3').Interior.Color = 4533685
    for ($i = 0; $i -lt $roles.Count; $i++) { $Sheet.Cells.Item($i + 4, 1).Value2 = $roles[$i] }
    for ($i = 0; $i -lt $spoSpecialties.Count; $i++) { $Sheet.Cells.Item($i + 4, 2).Value2 = $spoSpecialties[$i] }
    $Sheet.Range('E3:H5').Merge()
    $Sheet.Range('E3').Value2 = 'Справочники меняет координатор. Чтобы добавить специальность ВО или СПО, впишите её в свободную строку соответствующего столбца. Она появится в выпадающем списке.'
    $Sheet.Range('E3').WrapText = $true
    $Sheet.Range('E3:H5').Interior.Color = 15921906
    $Sheet.Range('A3:C50').Borders.LineStyle = 1
    $Sheet.Range('A3:C50').Borders.Color = 14277081
    $Sheet.Columns.Item(1).ColumnWidth = 37
    $Sheet.Columns.Item(2).ColumnWidth = 42
    $Sheet.Columns.Item(3).ColumnWidth = 42
    $Sheet.Columns.Item(5).ColumnWidth = 22
    $Sheet.Columns.Item(6).ColumnWidth = 18
    $Sheet.Columns.Item(7).ColumnWidth = 18
    $Sheet.Columns.Item(8).ColumnWidth = 18
}

$excel = New-Object -ComObject Excel.Application
$excel.Visible = $false
$excel.DisplayAlerts = $false
try {
    $book = $excel.Workbooks.Add(-4167)
    $spo = $book.Worksheets.Item(1)
    $spo.Name = 'СПО'
    $vo = $book.Worksheets.Add([Type]::Missing, $spo)
    $vo.Name = 'ВО'
    $settings = $book.Worksheets.Add([Type]::Missing, $vo)
    $settings.Name = 'Настройки'
    Set-SettingsSheet -Sheet $settings
    $book.Names.Add('Roles', "='Настройки'!`$A`$4:`$A`$50") | Out-Null
    $book.Names.Add('SPOSpecialties', "='Настройки'!`$B`$4:`$B`$50") | Out-Null
    $book.Names.Add('VOSpecialties', "='Настройки'!`$C`$4:`$C`$50") | Out-Null
    Set-RequestSheet -Sheet $spo -Profile 'СПО' -SpecialtyListName 'SPOSpecialties'
    Set-RequestSheet -Sheet $vo -Profile 'ВО' -SpecialtyListName 'VOSpecialties'
    $book.Worksheets.Item('СПО').Activate()
    $book.SaveAs($outputPath, 51)
    $book.Close($true)
    Write-Output $outputPath
}
finally {
    $excel.Quit()
}
