param(
  [Parameter(Mandatory = $true)][string]$Message,
  [Parameter(Mandatory = $true)][string]$Date
)

$env:GIT_AUTHOR_DATE = $Date
$env:GIT_COMMITTER_DATE = $Date
git -c user.name="nahouchimohamed" -c user.email="nahouchimohamed0@gmail.com" commit --author="nahouchimohamed <nahouchimohamed0@gmail.com>" -m $Message
Remove-Item Env:GIT_AUTHOR_DATE -ErrorAction SilentlyContinue
Remove-Item Env:GIT_COMMITTER_DATE -ErrorAction SilentlyContinue
