$ErrorActionPreference = 'Stop'

$outPath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot '..')) 'README.docx'
$readmePath = Join-Path (Resolve-Path (Join-Path $PSScriptRoot '..')) 'README.md'
if (Test-Path $outPath) { Remove-Item $outPath -Force }

$word = New-Object -ComObject Word.Application
$word.Visible = $false
$doc = $word.Documents.Add()
$selection = $word.Selection

function Add-Paragraph([string]$text, [int]$size = 11, [bool]$bold = $false) {
  $range = $selection.Range
  $range.Text = $text
  $range.Font.NameFarEast = 'Microsoft YaHei'
  $range.Font.Name = 'Microsoft YaHei'
  $range.Font.Size = $size
  $range.Font.Bold = [int]$bold
  $range.InsertParagraphAfter() | Out-Null
  $selection.MoveDown() | Out-Null
}

$lines = Get-Content -Path $readmePath -Encoding utf8
foreach ($line in $lines) {
  if ([string]::IsNullOrWhiteSpace($line)) {
    Add-Paragraph '' 11 $false
    continue
  }

  if ($line.StartsWith('# ')) {
    Add-Paragraph ($line.Substring(2)) 16 $true
    continue
  }

  if ($line.StartsWith('## ')) {
    Add-Paragraph ($line.Substring(3)) 13 $true
    continue
  }

  if ($line.StartsWith('- ')) {
    Add-Paragraph ('- ' + $line.Substring(2))
    continue
  }

  Add-Paragraph $line
}

$doc.SaveAs($outPath)
$doc.Close()
$word.Quit()

[System.Runtime.InteropServices.Marshal]::ReleaseComObject($selection) | Out-Null
[System.Runtime.InteropServices.Marshal]::ReleaseComObject($doc) | Out-Null
[System.Runtime.InteropServices.Marshal]::ReleaseComObject($word) | Out-Null

Write-Host "Created: $outPath"