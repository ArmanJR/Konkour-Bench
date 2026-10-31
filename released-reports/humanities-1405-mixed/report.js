// Lobe Icons 1.95.1 — https://github.com/lobehub/lobe-icons
// MIT License
//
// Copyright (c) 2023 LobeHub
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
const reportLogos = {"claude-color":"data:image/svg+xml;base64,PHN2ZyBoZWlnaHQ9IjFlbSIgc3R5bGU9ImZsZXg6bm9uZTtsaW5lLWhlaWdodDoxIiB2aWV3Qm94PSIwIDAgMjQgMjQiIHdpZHRoPSIxZW0iIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHRpdGxlPkNsYXVkZTwvdGl0bGU+PHBhdGggZD0iTTQuNzA5IDE1Ljk1NWw0LjcyLTIuNjQ3LjA4LS4yMy0uMDgtLjEyOEg5LjJsLS43OS0uMDQ4LTIuNjk4LS4wNzMtMi4zMzktLjA5Ny0yLjI2Ni0uMTIyLS41NzEtLjEyMUwwIDExLjc4NGwuMDU1LS4zNTIuNDgtLjMyMS42ODYuMDYgMS41Mi4xMDMgMi4yNzguMTU4IDEuNjUyLjA5NyAyLjQ0OS4yNTVoLjM4OWwuMDU1LS4xNTctLjEzNC0uMDk4LS4xMDMtLjA5Ny0yLjM1OC0xLjU5Ni0yLjU1Mi0xLjY4OC0xLjMzNi0uOTcyLS43MjQtLjQ5MS0uMzY0LS40NjItLjE1OC0xLjAwOC42NTYtLjcyMi44ODEuMDYuMjI1LjA2MS44OTMuNjg2IDEuOTA4IDEuNDc2IDIuNDkxIDEuODMzLjM2NS4zMDQuMTQ1LS4xMDMuMDE5LS4wNzMtLjE2NC0uMjc0LTEuMzU1LTIuNDQ2LTEuNDQ2LTIuNDktLjY0NC0xLjAzMi0uMTctLjYxOWEyLjk3IDIuOTcgMCAwMS0uMTA0LS43MjlMNi4yODMuMTM0IDYuNjk2IDBsLjk5Ni4xMzQuNDIuMzY0LjYyIDEuNDE0IDEuMDAyIDIuMjI5IDEuNTU1IDMuMDMuNDU2Ljg5OC4yNDMuODMyLjA5MS4yNTVoLjE1OFY5LjAxbC4xMjgtMS43MDYuMjM3LTIuMDk1LjIzLTIuNjk1LjA4LS43Ni4zNzYtLjkxLjc0Ny0uNDkyLjU4NC4yOC40OC42ODUtLjA2Ny40NDQtLjI4NiAxLjg1MS0uNTU5IDIuOTAzLS4zNjQgMS45NDJoLjIxMmwuMjQzLS4yNDIuOTg1LTEuMzA2IDEuNjUyLTIuMDY0LjczLS44Mi44NS0uOTA0LjU0Ny0uNDMxaDEuMDMzbC43NiAxLjEyOS0uMzQgMS4xNjYtMS4wNjQgMS4zNDctLjg4MSAxLjE0Mi0xLjI2NCAxLjctLjc5IDEuMzYuMDczLjExLjE4OC0uMDIgMi44NTYtLjYwNiAxLjU0My0uMjggMS44NDEtLjMxNS44MzMuMzg4LjA5MS4zOTUtLjMyOC44MDctMS45NjkuNDg2LTIuMzA5LjQ2Mi0zLjQzOS44MTMtLjA0Mi4wMy4wNDkuMDYxIDEuNTQ5LjE0Ni42NjIuMDM2aDEuNjIybDMuMDIuMjI1Ljc5LjUyMi40NzQuNjM4LS4wNzkuNDg1LTEuMjE1LjYyLTEuNjQtLjM4OS0zLjgyOS0uOTEtMS4zMTItLjMyOWgtLjE4MnYuMTFsMS4wOTMgMS4wNjggMi4wMDYgMS44MSAyLjUwOSAyLjMzLjEyNy41NzgtLjMyMi40NTUtLjM0LS4wNDktMi4yMDUtMS42NTctLjg1MS0uNzQ3LTEuOTI2LTEuNjJoLS4xMjh2LjE3bC40NDQuNjQ5IDIuMzQ1IDMuNTIxLjEyMiAxLjA4LS4xNy4zNTMtLjYwOC4yMTMtLjY2OC0uMTIyLTEuMzc0LTEuOTI1LTEuNDE1LTIuMTY3LTEuMTQzLTEuOTQzLS4xNC4wOC0uNjc0IDcuMjU0LS4zMTYuMzctLjcyOS4yOC0uNjA3LS40NjEtLjMyMi0uNzQ3LjMyMi0xLjQ3Ni4zODktMS45MjQuMzE1LTEuNTMuMjg2LTEuOS4xNy0uNjMyLS4wMTItLjA0Mi0uMTQuMDE4LTEuNDM0IDEuOTY3LTIuMTggMi45NDUtMS43MjYgMS44NDUtLjQxNC4xNjQtLjcxNy0uMzcuMDY3LS42NjIuNDAxLS41ODkgMi4zODgtMy4wMzYgMS40NC0xLjg4Mi45My0xLjA4Ni0uMDA2LS4xNThoLS4wNTVMNC4xMzIgMTguNTZsLTEuMTMuMTQ2LS40ODctLjQ1Ni4wNjEtLjc0Ni4yMzEtLjI0MyAxLjkwOC0xLjMxMi0uMDA2LjAwNnoiIGZpbGw9IiNEOTc3NTciIGZpbGwtcnVsZT0ibm9uemVybyI+PC9wYXRoPjwvc3ZnPg==","deepseek-color":"data:image/svg+xml;base64,PHN2ZyBoZWlnaHQ9IjFlbSIgc3R5bGU9ImZsZXg6bm9uZTtsaW5lLWhlaWdodDoxIiB2aWV3Qm94PSIwIDAgMjQgMjQiIHdpZHRoPSIxZW0iIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHRpdGxlPkRlZXBTZWVrPC90aXRsZT48cGF0aCBkPSJNMjMuNzQ4IDQuNDgyYy0uMjU0LS4xMjQtLjM2NC4xMTMtLjUxMi4yMzQtLjA1MS4wMzktLjA5NC4wOS0uMTM3LjEzNi0uMzcyLjM5Ny0uODA2LjY1Ny0xLjM3My42MjYtLjgyOS0uMDQ2LTEuNTM3LjIxNC0yLjE2My44NDgtLjEzMy0uNzgyLS41NzUtMS4yNDgtMS4yNDctMS41NDgtLjM1Mi0uMTU2LS43MDgtLjMxMS0uOTU1LS42NS0uMTcyLS4yNDEtLjIxOS0uNTEtLjMwNS0uNzc0LS4wNTUtLjE2LS4xMS0uMzIzLS4yOTMtLjM1LS4yLS4wMzEtLjI3OC4xMzYtLjM1Ni4yNzYtLjMxMy41NzItLjQzNCAxLjIwMi0uNDIyIDEuODQuMDI3IDEuNDM2LjYzMyAyLjU4IDEuODM4IDMuMzkzLjEzNy4wOTMuMTcyLjE4Ny4xMjkuMzIzLS4wODIuMjgtLjE4LjU1Mi0uMjY2LjgzMy0uMDU1LjE3OS0uMTM3LjIxNy0uMzI5LjE0YTUuNTI2IDUuNTI2IDAgMDEtMS43MzYtMS4xOGMtLjg1Ny0uODI4LTEuNjMxLTEuNzQyLTIuNTk3LTIuNDU4YTExLjM2NSAxMS4zNjUgMCAwMC0uNjg5LS40NzFjLS45ODUtLjk1Ny4xMy0xLjc0My4zODgtMS44MzYuMjctLjA5OC4wOTMtLjQzMi0uNzc5LS40MjgtLjg3Mi4wMDQtMS42Ny4yOTUtMi42ODcuNjg0YTMuMDU1IDMuMDU1IDAgMDEtLjQ2NS4xMzcgOS41OTcgOS41OTcgMCAwMC0yLjg4My0uMTAyYy0xLjg4NS4yMS0zLjM5IDEuMTAyLTQuNDk3IDIuNjIzQy4wODIgOC42MDYtLjIzMSAxMC42ODQuMTUyIDEyLjg1Yy40MDMgMi4yODQgMS41NjkgNC4xNzUgMy4zNiA1LjY1MyAxLjg1OCAxLjUzMyAzLjk5NyAyLjI4NCA2LjQzOCAyLjE0IDEuNDgyLS4wODUgMy4xMzMtLjI4NCA0Ljk5NC0xLjg2LjQ3LjIzNC45NjIuMzI3IDEuNzguMzk3LjYzLjA1OSAxLjIzNi0uMDMgMS43MDUtLjEyOC43MzUtLjE1Ni42ODQtLjgzNy40MTktLjk2MS0yLjE1NS0xLjAwNC0xLjY4Mi0uNTk1LTIuMTEzLS45MjYgMS4wOTYtMS4yOTYgMi43NDYtMi42NDIgMy4zOTItNy4wMDMuMDUtLjM0Ny4wMDctLjU2NSAwLS44NDUtLjAwNC0uMTcuMDM1LS4yMzcuMjMtLjI1NmE0LjE3MyA0LjE3MyAwIDAwMS41NDUtLjQ3NWMxLjM5Ni0uNzYzIDEuOTYtMi4wMTUgMi4wOTMtMy41MTcuMDItLjIzLS4wMDQtLjQ2Ny0uMjQ3LS41ODh6TTExLjU4MSAxOGMtMi4wODktMS42NDItMy4xMDItMi4xODMtMy41Mi0yLjE2LS4zOTIuMDI0LS4zMjEuNDcxLS4yMzUuNzYzLjA5LjI4OC4yMDcuNDg2LjM3MS43MzkuMTE0LjE2Ny4xOTIuNDE2LS4xMTMuNjAzLS42NzMuNDE2LTEuODQyLS4xNC0xLjg5Ny0uMTY3LTEuMzYxLS44MDItMi41LTEuODYtMy4zMDEtMy4zMDctLjc3NC0xLjM5My0xLjIyNC0yLjg4Ny0xLjI5OC00LjQ4Mi0uMDItLjM4Ni4wOTMtLjUyMi40NzctLjU5MmE0LjY5NiA0LjY5NiAwIDAxMS41MjktLjAzOWMyLjEzMi4zMTIgMy45NDYgMS4yNjUgNS40NjggMi43NzQuODY4Ljg2IDEuNTI1IDEuODg3IDIuMjAyIDIuODkxLjcyIDEuMDY2IDEuNDk0IDIuMDgyIDIuNDggMi45MTQuMzQ4LjI5Mi42MjUuNTE0Ljg5MS42NzctLjgwMi4wOS0yLjE0LjExLTMuMDU0LS42MTR6bTEtNi40NGEuMzA2LjMwNiAwIDAxLjQxNS0uMjg3LjMwMi4zMDIgMCAwMS4yLjI4OC4zMDYuMzA2IDAgMDEtLjMxLjMwNy4zMDMuMzAzIDAgMDEtLjMwNC0uMzA4em0zLjExIDEuNTk2Yy0uMi4wODEtLjM5OS4xNTEtLjU5LjE2YTEuMjQ1IDEuMjQ1IDAgMDEtLjc5OC0uMjU0Yy0uMjc0LS4yMy0uNDctLjM1OC0uNTUyLS43NThhMS43MyAxLjczIDAgMDEuMDE2LS41ODhjLjA3LS4zMjctLjAwOC0uNTM3LS4yMzktLjcyNy0uMTg3LS4xNTYtLjQyNi0uMTk5LS42ODgtLjE5OWEuNTU5LjU1OSAwIDAxLS4yNTQtLjA3OGMtLjExLS4wNTQtLjItLjE5LS4xMTQtLjM1OC4wMjgtLjA1NC4xNi0uMTg2LjE5Mi0uMjEuMzU2LS4yMDIuNzY3LS4xMzYgMS4xNDYuMDE2LjM1Mi4xNDQuNjE4LjQwOCAxLjAwMS43ODIuMzkxLjQ1MS40NjIuNTc2LjY4NS45MTQuMTc2LjI2NS4zMzYuNTM3LjQ0NS44NDguMDY3LjE5NS0uMDE5LjM1NC0uMjUuNDUyeiIgZmlsbD0iIzRENkJGRSI+PC9wYXRoPjwvc3ZnPg==","gemini-color":"data:image/svg+xml;base64,PHN2ZyBoZWlnaHQ9IjFlbSIgc3R5bGU9ImZsZXg6bm9uZTtsaW5lLWhlaWdodDoxIiB2aWV3Qm94PSIwIDAgMjQgMjQiIHdpZHRoPSIxZW0iIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHRpdGxlPkdlbWluaTwvdGl0bGU+PHBhdGggZD0iTTIwLjYxNiAxMC44MzVhMTQuMTQ3IDE0LjE0NyAwIDAxLTQuNDUtMy4wMDEgMTQuMTExIDE0LjExMSAwIDAxLTMuNjc4LTYuNDUyLjUwMy41MDMgMCAwMC0uOTc1IDAgMTQuMTM0IDE0LjEzNCAwIDAxLTMuNjc5IDYuNDUyIDE0LjE1NSAxNC4xNTUgMCAwMS00LjQ1IDMuMDAxYy0uNjUuMjgtMS4zMTguNTA1LTIuMDAyLjY3OGEuNTAyLjUwMiAwIDAwMCAuOTc1Yy42ODQuMTcyIDEuMzUuMzk3IDIuMDAyLjY3N2ExNC4xNDcgMTQuMTQ3IDAgMDE0LjQ1IDMuMDAxIDE0LjExMiAxNC4xMTIgMCAwMTMuNjc5IDYuNDUzLjUwMi41MDIgMCAwMC45NzUgMGMuMTcyLS42ODUuMzk3LTEuMzUxLjY3Ny0yLjAwM2ExNC4xNDUgMTQuMTQ1IDAgMDEzLjAwMS00LjQ1IDE0LjExMyAxNC4xMTMgMCAwMTYuNDUzLTMuNjc4LjUwMy41MDMgMCAwMDAtLjk3NSAxMy4yNDUgMTMuMjQ1IDAgMDEtMi4wMDMtLjY3OHoiIGZpbGw9IiMzMTg2RkYiPjwvcGF0aD48cGF0aCBkPSJNMjAuNjE2IDEwLjgzNWExNC4xNDcgMTQuMTQ3IDAgMDEtNC40NS0zLjAwMSAxNC4xMTEgMTQuMTExIDAgMDEtMy42NzgtNi40NTIuNTAzLjUwMyAwIDAwLS45NzUgMCAxNC4xMzQgMTQuMTM0IDAgMDEtMy42NzkgNi40NTIgMTQuMTU1IDE0LjE1NSAwIDAxLTQuNDUgMy4wMDFjLS42NS4yOC0xLjMxOC41MDUtMi4wMDIuNjc4YS41MDIuNTAyIDAgMDAwIC45NzVjLjY4NC4xNzIgMS4zNS4zOTcgMi4wMDIuNjc3YTE0LjE0NyAxNC4xNDcgMCAwMTQuNDUgMy4wMDEgMTQuMTEyIDE0LjExMiAwIDAxMy42NzkgNi40NTMuNTAyLjUwMiAwIDAwLjk3NSAwYy4xNzItLjY4NS4zOTctMS4zNTEuNjc3LTIuMDAzYTE0LjE0NSAxNC4xNDUgMCAwMTMuMDAxLTQuNDUgMTQuMTEzIDE0LjExMyAwIDAxNi40NTMtMy42NzguNTAzLjUwMyAwIDAwMC0uOTc1IDEzLjI0NSAxMy4yNDUgMCAwMS0yLjAwMy0uNjc4eiIgZmlsbD0idXJsKCNsb2JlLWljb25zLWdlbWluaS0wLV9SXzBfKSI+PC9wYXRoPjxwYXRoIGQ9Ik0yMC42MTYgMTAuODM1YTE0LjE0NyAxNC4xNDcgMCAwMS00LjQ1LTMuMDAxIDE0LjExMSAxNC4xMTEgMCAwMS0zLjY3OC02LjQ1Mi41MDMuNTAzIDAgMDAtLjk3NSAwIDE0LjEzNCAxNC4xMzQgMCAwMS0zLjY3OSA2LjQ1MiAxNC4xNTUgMTQuMTU1IDAgMDEtNC40NSAzLjAwMWMtLjY1LjI4LTEuMzE4LjUwNS0yLjAwMi42NzhhLjUwMi41MDIgMCAwMDAgLjk3NWMuNjg0LjE3MiAxLjM1LjM5NyAyLjAwMi42NzdhMTQuMTQ3IDE0LjE0NyAwIDAxNC40NSAzLjAwMSAxNC4xMTIgMTQuMTEyIDAgMDEzLjY3OSA2LjQ1My41MDIuNTAyIDAgMDAuOTc1IDBjLjE3Mi0uNjg1LjM5Ny0xLjM1MS42NzctMi4wMDNhMTQuMTQ1IDE0LjE0NSAwIDAxMy4wMDEtNC40NSAxNC4xMTMgMTQuMTEzIDAgMDE2LjQ1My0zLjY3OC41MDMuNTAzIDAgMDAwLS45NzUgMTMuMjQ1IDEzLjI0NSAwIDAxLTIuMDAzLS42Nzh6IiBmaWxsPSJ1cmwoI2xvYmUtaWNvbnMtZ2VtaW5pLTEtX1JfMF8pIj48L3BhdGg+PHBhdGggZD0iTTIwLjYxNiAxMC44MzVhMTQuMTQ3IDE0LjE0NyAwIDAxLTQuNDUtMy4wMDEgMTQuMTExIDE0LjExMSAwIDAxLTMuNjc4LTYuNDUyLjUwMy41MDMgMCAwMC0uOTc1IDAgMTQuMTM0IDE0LjEzNCAwIDAxLTMuNjc5IDYuNDUyIDE0LjE1NSAxNC4xNTUgMCAwMS00LjQ1IDMuMDAxYy0uNjUuMjgtMS4zMTguNTA1LTIuMDAyLjY3OGEuNTAyLjUwMiAwIDAwMCAuOTc1Yy42ODQuMTcyIDEuMzUuMzk3IDIuMDAyLjY3N2ExNC4xNDcgMTQuMTQ3IDAgMDE0LjQ1IDMuMDAxIDE0LjExMiAxNC4xMTIgMCAwMTMuNjc5IDYuNDUzLjUwMi41MDIgMCAwMC45NzUgMGMuMTcyLS42ODUuMzk3LTEuMzUxLjY3Ny0yLjAwM2ExNC4xNDUgMTQuMTQ1IDAgMDEzLjAwMS00LjQ1IDE0LjExMyAxNC4xMTMgMCAwMTYuNDUzLTMuNjc4LjUwMy41MDMgMCAwMDAtLjk3NSAxMy4yNDUgMTMuMjQ1IDAgMDEtMi4wMDMtLjY3OHoiIGZpbGw9InVybCgjbG9iZS1pY29ucy1nZW1pbmktMi1fUl8wXykiPjwvcGF0aD48ZGVmcz48bGluZWFyR3JhZGllbnQgZ3JhZGllbnRVbml0cz0idXNlclNwYWNlT25Vc2UiIGlkPSJsb2JlLWljb25zLWdlbWluaS0wLV9SXzBfIiB4MT0iNyIgeDI9IjExIiB5MT0iMTUuNSIgeTI9IjEyIj48c3RvcCBzdG9wLWNvbG9yPSIjMDhCOTYyIj48L3N0b3A+PHN0b3Agb2Zmc2V0PSIxIiBzdG9wLWNvbG9yPSIjMDhCOTYyIiBzdG9wLW9wYWNpdHk9IjAiPjwvc3RvcD48L2xpbmVhckdyYWRpZW50PjxsaW5lYXJHcmFkaWVudCBncmFkaWVudFVuaXRzPSJ1c2VyU3BhY2VPblVzZSIgaWQ9ImxvYmUtaWNvbnMtZ2VtaW5pLTEtX1JfMF8iIHgxPSI4IiB4Mj0iMTEuNSIgeTE9IjUuNSIgeTI9IjExIj48c3RvcCBzdG9wLWNvbG9yPSIjRjk0NTQzIj48L3N0b3A+PHN0b3Agb2Zmc2V0PSIxIiBzdG9wLWNvbG9yPSIjRjk0NTQzIiBzdG9wLW9wYWNpdHk9IjAiPjwvc3RvcD48L2xpbmVhckdyYWRpZW50PjxsaW5lYXJHcmFkaWVudCBncmFkaWVudFVuaXRzPSJ1c2VyU3BhY2VPblVzZSIgaWQ9ImxvYmUtaWNvbnMtZ2VtaW5pLTItX1JfMF8iIHgxPSIzLjUiIHgyPSIxNy41IiB5MT0iMTMuNSIgeTI9IjEyIj48c3RvcCBzdG9wLWNvbG9yPSIjRkFCQzEyIj48L3N0b3A+PHN0b3Agb2Zmc2V0PSIuNDYiIHN0b3AtY29sb3I9IiNGQUJDMTIiIHN0b3Atb3BhY2l0eT0iMCI+PC9zdG9wPjwvbGluZWFyR3JhZGllbnQ+PC9kZWZzPjwvc3ZnPg==","grok":"data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiIGZpbGwtcnVsZT0iZXZlbm9kZCIgaGVpZ2h0PSIxZW0iIHN0eWxlPSJmbGV4Om5vbmU7bGluZS1oZWlnaHQ6MSIgdmlld0JveD0iMCAwIDI0IDI0IiB3aWR0aD0iMWVtIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjx0aXRsZT5Hcm9rPC90aXRsZT48cGF0aCBkPSJNOS4yNyAxNS4yOWw3Ljk3OC01Ljg5N2MuMzkxLS4yOS45NS0uMTc3IDEuMTM3LjI3Mi45OCAyLjM2OS41NDIgNS4yMTUtMS40MSA3LjE2OS0xLjk1MSAxLjk1NC00LjY2NyAyLjM4Mi03LjE0OSAxLjQwNmwtMi43MTEgMS4yNTdjMy44ODkgMi42NjEgOC42MTEgMi4wMDMgMTEuNTYyLS45NTMgMi4zNDEtMi4zNDQgMy4wNjYtNS41MzkgMi4zODgtOC40MmwuMDA2LjAwN2MtLjk4My00LjIzMi4yNDItNS45MjQgMi43NS05LjM4My4wNi0uMDgyLjEyLS4xNjQuMTc5LS4yNDhsLTMuMzAxIDMuMzA1di0uMDFMOS4yNjcgMTUuMjkyTTcuNjIzIDE2LjcyM2MtMi43OTItMi42Ny0yLjMxLTYuODAxLjA3MS05LjE4NCAxLjc2MS0xLjc2MyA0LjY0Ny0yLjQ4MyA3LjE2Ni0xLjQyNWwyLjcwNS0xLjI1YTcuODA4IDcuODA4IDAgMDAtMS44MjktMUE4Ljk3NSA4Ljk3NSAwIDAwNS45ODQgNS44M2MtMi41MzMgMi41MzYtMy4zMyA2LjQzNi0xLjk2MiA5Ljc2NCAxLjAyMiAyLjQ4Ny0uNjUzIDQuMjQ2LTIuMzQgNi4wMjItLjU5OS42My0xLjE5OSAxLjI1OS0xLjY4MiAxLjkyNWw3LjYyLTYuODE1Ij48L3BhdGg+PC9zdmc+","openai":"data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiIGZpbGwtcnVsZT0iZXZlbm9kZCIgaGVpZ2h0PSIxZW0iIHN0eWxlPSJmbGV4Om5vbmU7bGluZS1oZWlnaHQ6MSIgdmlld0JveD0iMCAwIDI0IDI0IiB3aWR0aD0iMWVtIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjx0aXRsZT5PcGVuQUk8L3RpdGxlPjxwYXRoIGQ9Ik05LjIwNSA4LjY1OHYtMi4yNmMwLS4xOS4wNzItLjMzMy4yMzgtLjQyOGw0LjU0My0yLjYxNmMuNjE5LS4zNTcgMS4zNTYtLjUyMyAyLjExNy0uNTIzIDIuODU0IDAgNC42NjIgMi4yMTIgNC42NjIgNC41NjYgMCAuMTY3IDAgLjM1Ny0uMDI0LjU0N2wtNC43MS0yLjc1OWEuNzk3Ljc5NyAwIDAwLS44NTYgMGwtNS45NyAzLjQ3M3ptMTAuNjA5IDguOFYxMi4wNmMwLS4zMzMtLjE0My0uNTctLjQyOS0uNzM3bC01Ljk3LTMuNDczIDEuOTUtMS4xMThhLjQzMy40MzMgMCAwMS40NzYgMGw0LjU0MyAyLjYxN2MxLjMwOS43NiAyLjE4OSAyLjM3OCAyLjE4OSAzLjk0OCAwIDEuODA4LTEuMDcgMy40NzMtMi43NiA0LjE2M3pNNy44MDIgMTIuNzAzbC0xLjk1LTEuMTQyYy0uMTY3LS4wOTUtLjIzOS0uMjM4LS4yMzktLjQyOFY1Ljg5OWMwLTIuNTQ1IDEuOTUtNC40NzIgNC41OTEtNC40NzIgMSAwIDEuOTI3LjMzMyAyLjcxMi45MjhMOC4yMyA1LjA2N2MtLjI4NS4xNjYtLjQyOC40MDQtLjQyOC43Mzd2Ni44OTh6TTEyIDE1LjEyOGwtMi43OTUtMS41N3YtMy4zM0wxMiA4LjY1OGwyLjc5NSAxLjU3djMuMzNMMTIgMTUuMTI4em0xLjc5NiA3LjIzYy0xIDAtMS45MjctLjMzMi0yLjcxMi0uOTI3bDQuNjg2LTIuNzEyYy4yODUtLjE2Ni40MjgtLjQwNC40MjgtLjczN3YtNi44OThsMS45NzQgMS4xNDJjLjE2Ny4wOTUuMjM4LjIzOC4yMzguNDI4djUuMjMzYzAgMi41NDUtMS45NzQgNC40NzItNC42MTQgNC40NzJ6bS01LjYzNy01LjMwM2wtNC41NDQtMi42MTdjLTEuMzA4LS43NjEtMi4xODgtMi4zNzgtMi4xODgtMy45NDhBNC40ODIgNC40ODIgMCAwMTQuMjEgNi4zMjd2NS40MjNjMCAuMzMzLjE0My41NzEuNDI4LjczOGw1Ljk0NyAzLjQ0OS0xLjk1IDEuMTE4YS40MzIuNDMyIDAgMDEtLjQ3NiAwem0tLjI2MiAzLjljLTIuNjg4IDAtNC42NjItMi4wMjEtNC42NjItNC41MTkgMC0uMTkuMDI0LS4zOC4wNDctLjU3bDQuNjg2IDIuNzFjLjI4Ni4xNjcuNTcxLjE2Ny44NTYgMGw1Ljk3LTMuNDQ4djIuMjZjMCAuMTktLjA3LjMzMy0uMjM3LjQyOGwtNC41NDMgMi42MTZjLS42MTkuMzU3LTEuMzU2LjUyMy0yLjExNy41MjN6bTUuODk5IDIuODNhNS45NDcgNS45NDcgMCAwMDUuODI3LTQuNzU2QzIyLjI4NyAxOC4zMzkgMjQgMTUuODQgMjQgMTMuMjk2YzAtMS42NjUtLjcxMy0zLjI4Mi0xLjk5OC00LjQ0OC4xMTktLjUuMTktLjk5OS4xOS0xLjQ5OCAwLTMuNDAxLTIuNzU5LTUuOTQ3LTUuOTQ2LTUuOTQ3LS42NDIgMC0xLjI2LjA5NS0xLjg4LjMxQTUuOTYyIDUuOTYyIDAgMDAxMC4yMDUgMGE1Ljk0NyA1Ljk0NyAwIDAwLTUuODI3IDQuNzU3QzEuNzEzIDUuNDQ3IDAgNy45NDUgMCAxMC40OWMwIDEuNjY2LjcxMyAzLjI4MyAxLjk5OCA0LjQ0OC0uMTE5LjUtLjE5IDEtLjE5IDEuNDk5IDAgMy40MDEgMi43NTkgNS45NDYgNS45NDYgNS45NDYuNjQyIDAgMS4yNi0uMDk1IDEuODgtLjMwOWE1Ljk2IDUuOTYgMCAwMDQuMTYyIDEuNzEzeiI+PC9wYXRoPjwvc3ZnPg==","qwen-color":"data:image/svg+xml;base64,PHN2ZyBoZWlnaHQ9IjFlbSIgc3R5bGU9ImZsZXg6bm9uZTtsaW5lLWhlaWdodDoxIiB2aWV3Qm94PSIwIDAgMjQgMjQiIHdpZHRoPSIxZW0iIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHRpdGxlPlF3ZW48L3RpdGxlPjxwYXRoIGQ9Ik0xMi42MDQgMS4zNGMuMzkzLjY5Ljc4NCAxLjM4MiAxLjE3NCAyLjA3NWEuMTguMTggMCAwMC4xNTcuMDkxaDUuNTUyYy4xNzQgMCAuMzIyLjExLjQ0Ni4zMjdsMS40NTQgMi41N2MuMTkuMzM3LjI0LjQ3OC4wMjQuODM3LS4yNi40My0uNTEzLjg2NC0uNzYgMS4zbC0uMzY3LjY1OGMtLjEwNi4xOTYtLjIyMy4yOC0uMDQuNTEybDIuNjUyIDQuNjM3Yy4xNzIuMzAxLjExMS40OTQtLjA0My43Ny0uNDM3Ljc4NS0uODgyIDEuNTY0LTEuMzM1IDIuMzQtLjE1OS4yNzItLjM1Mi4zNzUtLjY4LjM3LS43NzctLjAxNi0xLjU1Mi0uMDEtMi4zMjcuMDE2YS4wOTkuMDk5IDAgMDAtLjA4MS4wNSA1NzUuMDk3IDU3NS4wOTcgMCAwMS0yLjcwNSA0Ljc0Yy0uMTY5LjI5My0uMzguMzYzLS43MjUuMzY0LS45OTcuMDAzLTIuMDAyLjAwNC0zLjAxNy4wMDJhLjUzNy41MzcgMCAwMS0uNDY1LS4yNzFsLTEuMzM1LTIuMzIzYS4wOS4wOSAwIDAwLS4wODMtLjA0OUg0Ljk4MmMtLjI4NS4wMy0uNTUzLS4wMDEtLjgwNS0uMDkybC0xLjYwMy0yLjc3YS41NDMuNTQzIDAgMDEtLjAwMi0uNTRsMS4yMDctMi4xMmEuMTk4LjE5OCAwIDAwMC0uMTk3IDU1MC45NTEgNTUwLjk1MSAwIDAxLTEuODc1LTMuMjcybC0uNzktMS4zOTVjLS4xNi0uMzEtLjE3My0uNDk2LjA5NS0uOTY1LjQ2NS0uODEzLjkyNy0xLjYyNSAxLjM4Ny0yLjQzNi4xMzItLjIzNC4zMDQtLjMzNC41ODQtLjMzNWEzMzguMyAzMzguMyAwIDAxMi41ODktLjAwMS4xMjQuMTI0IDAgMDAuMTA3LS4wNjNsMi44MDYtNC44OTVhLjQ4OC40ODggMCAwMS40MjItLjI0NmMuNTI0LS4wMDEgMS4wNTMgMCAxLjU4My0uMDA2TDExLjcwNCAxYy4zNDEtLjAwMy43MjQuMDMyLjkuMzR6bS0zLjQzMi40MDNhLjA2LjA2IDAgMDAtLjA1Mi4wM0w2LjI1NCA2Ljc4OGEuMTU3LjE1NyAwIDAxLS4xMzUuMDc4SDMuMjUzYy0uMDU2IDAtLjA3LjAyNS0uMDQxLjA3NGw1LjgxIDEwLjE1NmMuMDI1LjA0Mi4wMTMuMDYyLS4wMzQuMDYzbC0yLjc5NS4wMTVhLjIxOC4yMTggMCAwMC0uMi4xMTZsLTEuMzIgMi4zMWMtLjA0NC4wNzgtLjAyMS4xMTguMDY4LjExOGw1LjcxNi4wMDhjLjA0NiAwIC4wOC4wMi4xMDQuMDYxbDEuNDAzIDIuNDU0Yy4wNDYuMDgxLjA5Mi4wODIuMTM5IDBsNS4wMDYtOC43Ni43ODMtMS4zODJhLjA1NS4wNTUgMCAwMS4wOTYgMGwxLjQyNCAyLjUzYS4xMjIuMTIyIDAgMDAuMTA3LjA2MmwyLjc2My0uMDJhLjA0LjA0IDAgMDAuMDM1LS4wMi4wNDEuMDQxIDAgMDAwLS4wNGwtMi45LTUuMDg2YS4xMDguMTA4IDAgMDEwLS4xMTNsLjI5My0uNTA3IDEuMTItMS45NzdjLjAyNC0uMDQxLjAxMi0uMDYyLS4wMzUtLjA2Mkg5LjJjLS4wNTkgMC0uMDczLS4wMjYtLjA0My0uMDc3bDEuNDM0LTIuNTA1YS4xMDcuMTA3IDAgMDAwLS4xMTRMOS4yMjUgMS43NzRhLjA2LjA2IDAgMDAtLjA1My0uMDMxem02LjI5IDguMDJjLjA0NiAwIC4wNTguMDIuMDM0LjA2bC0uODMyIDEuNDY1LTIuNjEzIDQuNTg1YS4wNTYuMDU2IDAgMDEtLjA1LjAyOS4wNTguMDU4IDAgMDEtLjA1LS4wMjlMOC40OTggOS44NDFjLS4wMi0uMDM0LS4wMS0uMDUyLjAyOC0uMDU0bC4yMTYtLjAxMiA2LjcyMi0uMDEyeiIgZmlsbD0idXJsKCNsb2JlLWljb25zLXF3ZW4tX1JfMF8pIiBmaWxsLXJ1bGU9Im5vbnplcm8iPjwvcGF0aD48ZGVmcz48bGluZWFyR3JhZGllbnQgaWQ9ImxvYmUtaWNvbnMtcXdlbi1fUl8wXyIgeDE9IjAlIiB4Mj0iMTAwJSIgeTE9IjAlIiB5Mj0iMCUiPjxzdG9wIG9mZnNldD0iMCUiIHN0b3AtY29sb3I9IiM2MzM2RTciIHN0b3Atb3BhY2l0eT0iLjg0Ij48L3N0b3A+PHN0b3Agb2Zmc2V0PSIxMDAlIiBzdG9wLWNvbG9yPSIjNkY2OUY3IiBzdG9wLW9wYWNpdHk9Ii44NCI+PC9zdG9wPjwvbGluZWFyR3JhZGllbnQ+PC9kZWZzPjwvc3ZnPg==","xiaomimimo":"data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiIGZpbGwtcnVsZT0iZXZlbm9kZCIgaGVpZ2h0PSIxZW0iIHN0eWxlPSJmbGV4Om5vbmU7bGluZS1oZWlnaHQ6MSIgdmlld0JveD0iMCAwIDI0IDI0IiB3aWR0aD0iMWVtIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjx0aXRsZT5YaWFvbWlNaU1vPC90aXRsZT48cGF0aCBkPSJNLjk1OCAxNS45MzZhLjQ1OS40NTkgMCAwMS40NTkuNDR2Mi43MjlhLjQ2LjQ2IDAgMDEtLjkxOCAwdi0yLjcyOWEuNDU5LjQ1OSAwIDAxLjQ1OS0uNDR6bTQuODE0LTIuMDM1YS40Ni40NiAwIDAxLjU1My40NXY0Ljc1NGEuNDU4LjQ1OCAwIDExLS45MTggMFYxNS40OEwzLjc0IDE3LjIwMmEuNDYyLjQ2MiAwIDAxLS42NTUuMDE2LjQ2Mi40NjIgMCAwMS0uMDY1LS4wODJMLjYyOCAxNC42N2EuNDU5LjQ1OSAwIDAxLjY1OC0uNjM3bDIuMTI0IDIuMTg3IDIuMTI3LTIuMTg4YS40Ni40NiAwIDAxLjIzNS0uMTN6bTIuMDY4LjAwNGEuNDYuNDYgMCAwMS40NTguNDQ1djQuNzU1YS40Ni40NiAwIDAxLS40NTguNDU4LjQ1OS40NTkgMCAwMS0uNDU4LS40NThWMTQuMzVhLjQ1OS40NTkgMCAwMS40NTgtLjQ0NXptMS45NzMgMi4wMTRhLjQ2LjQ2IDAgMDEuNDYuNDU3djIuNzI5YS40Ni40NiAwIDAxLS43ODQuMzI0LjQ2LjQ2IDAgMDEtLjEzNC0uMzI0di0yLjcyOWEuNDYuNDYgMCAwMS40NTgtLjQ1OHptLjAwMi0yLjA0NWEuNDU4LjQ1OCAwIDAxLjMyOC4xNTdsMi4xMjcgMi4xOSAyLjEyNS0yLjE5YS40NTkuNDU5IDAgMDEuNzg0LjMxOHY0Ljc1NmEuNDYuNDYgMCAwMS0uNDU1LjQ1OC40Ni40NiAwIDAxLS40NTgtLjQ1OFYxNS40OGwtMS42NjcgMS43MjNhLjQ2LjQ2IDAgMDEtLjY1LjAwOGwtLjAwNS0uMDA1YzAtLjAwMi0uMDAyLS4wMDItLjAwNC0uMDAzbC0yLjQ1NS0yLjUzNGEuNDYuNDYgMCAwMS0uMDA4LS42NjcuNDYxLjQ2MSAwIDAxLjMzOC0uMTI4em02Ljc5NyAxLjIwNmEuNDYuNDYgMCAwMS41My42NTFBMS45NjYgMS45NjYgMCAwMDE5LjgxIDE4LjRhLjQ2Mi40NjIgMCAwMS42MjMuMTguNDYuNDYgMCAwMS0uMTgxLjYyNCAyLjg2MyAyLjg2MyAwIDAxLTEuMzguMzUzbC0uMTQyLS4wMDRhMi44OCAyLjg4IDAgMDEtMi4zOTMtNC4yNjMuNDYxLjQ2MSAwIDAxLjI3NC0uMjF6bS44NjQtLjkzMWEyLjg4NCAyLjg4NCAwIDAxMy45MTUgMy45MTQuNDYuNDYgMCAwMS0uNDAyLjI0bC0uMDU3LS4wMDRhLjQ1OC40NTggMCAwMS0uMTY0LS4wNTUuNDYuNDYgMCAwMS0uMTgyLS42MjIgMS45NjcgMS45NjcgMCAwMC0yLjY2OS0yLjY3LjQ1OS40NTkgMCAxMS0uNDQxLS44MDN6TTkuNTkgNi4zNjhjMS40ODEgMCAxLjY5NiAxLjIwMiAxLjY5NiAxLjY1NHYyLjY0OGgtLjkxN3YtLjQzMmMtLjI2LjM0Ni0uNzkyLjUzNS0xLjM2LjUzNS0uMTMzIDAtMS4yODktLjAzLTEuMzg0LTEuMTM2LS4wODItLjkzMi42NzUtMS42MSAyLjA1My0xLjYxaC42OTFjMC0uNTYzLS4zNjctLjg4Ni0uOTgzLS44ODYtLjQ0LjAxMy0uODY0LjE3NC0xLjIuNDU4bC0uMzYtLjY2NGMuNDg0LS4zNzkgMS4wMTItLjU2NyAxLjc2NC0uNTY3em00LjQyNy4xYzEuMjYzIDAgMi4wODIuOTcgMi4wODMgMi4xNSAwIDEuMTgxLS44MjQgMi4xNTQtMi4wODMgMi4xNTQtMS4yNiAwLTIuMDg0LS45NzItMi4wODQtMi4xNTIgMC0xLjE4LjgyLTIuMTUzIDIuMDg0LTIuMTUzem02LjgwMS4wMTVjLjY4IDAgMS4yMDIuNDY1IDEuMTk3IDEuNTQ4djIuNjQySDIxLjFWOC4yOWMwLS4zMTItLjAwMi0uOTgtLjYzLS45OHMtLjYyOC42NjctLjYyOC44Mzh2Mi41MjRoLS44OVY4LjE0OGMwLS4xNy0uMDAxLS44MzgtLjYzLS44MzgtLjYyOCAwLS42MjguNjY4LS42MjguOTh2Mi4zODNoLS45MTd2LTQuMDNoLjkxN1Y3YTEuMjIgMS4yMiAwIDAxLjk0Ny0uNTE2Yy4zOTggMCAuNzYuMTkzLjk4Mi42ODZhMS4zMjEgMS4zMjEgMCAwMTEuMTk1LS42ODZ6bS0xOC4wOTMuODcybDEuNDU3LTEuNzcySDUuMzJMMy4zMTEgOC4wN2wyLjE0IDIuNjAySDQuMjRMMi43MjUgOC43OTYgMS4yMSAxMC42NzJIMEwyLjEzOCA4LjA3LjEzIDUuNTgzaDEuMTM4bDEuNDU4IDEuNzcyem00LjE0OSAzLjMxN2gtLjkxNlY2LjY0NGguOTE2djQuMDI4em0xNi45OSAwaC0uOTE2VjYuNjQ0aC45MTZ2NC4wMjh6TTkuOTI1IDguNzFjLTEuMDU1IDAtMS4zNTkuNDEyLTEuMzI2Ljc0Mi4wMzIuMzI5LjMyNC41MzcuNzU3LjUzN2ExLjAxMyAxLjAxMyAwIDAwMS4wMTQtLjk2OGwuMDAyLS4zMWgtLjQ0N3pNMTQuMDE4IDcuM2MtLjY2MyAwLTEuMTg0LjQ4Ny0xLjE4NCAxLjMyIDAgLjgzMi41MiAxLjMyIDEuMTg0IDEuMzIuNjYyIDAgMS4xODItLjQ5IDEuMTgyLTEuMzIgMC0uODMyLS41Mi0xLjMyLTEuMTgyLTEuMzJ6TTYuNDE3IDUuMDAxYS41NjguNTY4IDAgMDEuNTg3LjU4Mi41ODguNTg4IDAgMDEtMS4xNzUgMEEuNTcuNTcgMCAwMTYuNDE3IDV6bTE2Ljk5MSAwYS41Ny41NyAwIDAxLjU5Mi41ODIuNTg4LjU4OCAwIDAxLTEuMTc0IDAgLjU3LjU3IDAgMDEuMzU3LS41NDIuNTcyLjU3MiAwIDAxLjIyNS0uMDR6Ij48L3BhdGg+PC9zdmc+","zai":"data:image/svg+xml;base64,PHN2ZyBmaWxsPSJjdXJyZW50Q29sb3IiIGZpbGwtcnVsZT0iZXZlbm9kZCIgaGVpZ2h0PSIxZW0iIHN0eWxlPSJmbGV4Om5vbmU7bGluZS1oZWlnaHQ6MSIgdmlld0JveD0iMCAwIDI0IDI0IiB3aWR0aD0iMWVtIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjx0aXRsZT5aLmFpPC90aXRsZT48cGF0aCBkPSJNMTIuMTA1IDJMOS45MjcgNC45NTNILjY1M0wyLjgzIDJoOS4yNzZ6TTIzLjI1NCAxOS4wNDhMMjEuMDc4IDIyaC05LjI0MmwyLjE3NC0yLjk1Mmg5LjI0NHpNMjQgMkw5LjI2NCAyMkgwTDE0LjczNiAySDI0eiI+PC9wYXRoPjwvc3ZnPg=="};
"use strict";

// Display typography only; canonical IDs continue to identify every result.
const modelDisplayNames = new Map([
  ["anthropic/claude-opus-5.5", "Claude Opus 5.5"],
  ["anthropic/claude-sonnet-5.5", "Claude Sonnet 5.5"],
  ["google/gemini-3.8-flash", "Gemini 3.8 Flash"],
  ["google/gemini-4-argon", "Gemini 4 Argon"],
  ["openai/gpt-6.1-sol", "GPT-6.1 Sol"],
  ["openai/gpt-6-luna", "GPT-6 Luna"],
  ["deepseek/deepseek-v4.1-flash-1017", "DeepSeek V4.1 Flash 1017"],
  ["deepseek/deepseek-v4-pro-0813", "DeepSeek V4 Pro 0813"],
  ["qwen/qwen3.8-27b", "Qwen 3.8 27B"],
  ["qwen/qwen3.8-max-0902", "Qwen 3.8 Max 0902"],
  ["x-ai/grok-4.7", "Grok 4.7"],
  ["xiaomi/mimo-v2.6-pro", "MiMo V2.6 Pro"],
  ["xiaomi/mimo-v2.6-flash", "MiMo V2.6 Flash"],
  ["z-ai/glm-5.3", "GLM 5.3"],
  ["z-ai/glm-5.3-flash", "GLM 5.3 Flash"],
]);

function modelDisplayName(model) {
  return (
    modelDisplayNames.get(model) || model.split("/").slice(1).join("/") || model
  );
}

function modelMatchesQuery(model, query) {
  return `${model} ${modelDisplayName(model)}`
    .toLowerCase()
    .includes(query.toLowerCase().trim());
}

// Match model families, not arbitrary substrings or the inference backend.
function modelBrand(model) {
  if (typeof model !== "string") return null;
  const families = [
    ["openai/", "openai"],
    ["anthropic/claude-", "claude-color"],
    ["google/gemini-", "gemini-color"],
    ["deepseek/", "deepseek-color"],
    ["qwen/", "qwen-color"],
    ["x-ai/grok-", "grok"],
    ["xiaomi/mimo-", "xiaomimimo"],
    ["z-ai/glm-", "zai"],
  ];
  return families.find(([prefix]) => model.startsWith(prefix))?.[1] || null;
}

function modelLogo(
  model,
  logos = typeof reportLogos === "undefined" ? {} : reportLogos,
) {
  const brand = modelBrand(model);
  return brand ? logos[brand] || "" : "";
}

if (typeof module !== "undefined" && module.exports)
  module.exports = {
    modelBrand,
    modelLogo,
    modelDisplayName,
    modelMatchesQuery,
  };

"use strict";

function countAxisMaximum(max) {
  return Math.max(4, Math.ceil(max / 4) * 4);
}

function heatmapColor(value) {
  const rgb = [239 - value * 231, 246 - value * 129, 241 - value * 137].map(
    Math.round,
  );
  const linear = rgb.map((n) => {
    const s = n / 255;
    return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
  });
  const luminance =
    linear[0] * 0.2126 + linear[1] * 0.7152 + linear[2] * 0.0722;
  return {
    fill: `rgb(${rgb.join(",")})`,
    text: luminance > 0.179 ? "#000" : "#fff",
  };
}

// Bound canvas allocations without cropping tall figures or changing aspect ratio.
function pngDimensions(width, height) {
  if (!(width > 0 && height > 0 && Number.isFinite(width + height)))
    throw new Error("The chart has invalid dimensions.");
  const scale = Math.min(
    2,
    16384 / width,
    16384 / height,
    Math.sqrt(32000000 / (width * height)),
  );
  return {
    width: Math.max(1, Math.floor(width * scale)),
    height: Math.max(1, Math.floor(height * scale)),
  };
}

// Place compact labels next to their own data points. Candidate positions are
// scored against markers, the frontier, and other labels; data never moves.
function frontierLabelPositions(labels, bounds, points, segments = []) {
  const overlap = (a, b, pad = 6) =>
    Math.max(
      0,
      Math.min(a.x + a.width + pad, b.x + b.width) - Math.max(a.x - pad, b.x),
    ) *
    Math.max(
      0,
      Math.min(a.y + a.height + pad, b.y + b.height) - Math.max(a.y - pad, b.y),
    );
  const crosses = (box, [a, b]) => {
    let enter = 0,
      leave = 1;
    for (const [axis, extent] of [
      ["x", "width"],
      ["y", "height"],
    ]) {
      const delta = b[axis] - a[axis];
      const low = box[axis] - 4,
        high = box[axis] + box[extent] + 4;
      if (delta === 0) {
        if (a[axis] < low || a[axis] > high) return false;
      } else {
        const t1 = (low - a[axis]) / delta,
          t2 = (high - a[axis]) / delta;
        enter = Math.max(enter, Math.min(t1, t2));
        leave = Math.min(leave, Math.max(t1, t2));
        if (enter > leave) return false;
      }
    }
    return true;
  };
  const candidates = labels.map(({ x, y, width, height }) => {
    if (
      width > bounds.right - bounds.left ||
      height > bounds.bottom - bounds.top
    )
      throw new Error("The frontier label is larger than the plot.");
    return [16, 32, 48].flatMap((gap) =>
      [
        [x + gap, y - height - gap],
        [x - width - gap, y - height - gap],
        [x + gap, y + gap],
        [x - width - gap, y + gap],
        [x - width / 2, y - height - gap],
        [x - width / 2, y + gap],
        [x + gap, y - height / 2],
        [x - width - gap, y - height / 2],
      ].map(([left, top], preference) => {
        const box = {
          x: Math.max(bounds.left, Math.min(bounds.right - width, left)),
          y: Math.max(bounds.top, Math.min(bounds.bottom - height, top)),
          width,
          height,
        };
        const distance = Math.hypot(
          Math.max(box.x - x, 0, x - box.x - width),
          Math.max(box.y - y, 0, y - box.y - height),
        );
        const collisions = points.filter((p) =>
          overlap(
            box,
            {
              x: p.x - 10,
              y: p.y - 10,
              width: 20,
              height: 20,
            },
            3,
          ),
        ).length;
        return {
          ...box,
          penalty:
            collisions * 100000 +
            segments.filter((line) => crosses(box, line)).length * 10000 +
            distance +
            preference * 2,
        };
      }),
    );
  });
  const placed = [];
  // A few deterministic relaxation passes resolve close frontier neighbours.
  for (let pass = 0; pass < 4; pass++) {
    candidates.forEach((options, i) => {
      let best,
        bestScore = Infinity;
      for (const option of options) {
        const score =
          option.penalty +
          placed.reduce(
            (sum, other, j) =>
              sum + (i === j ? 0 : overlap(option, other) * 1000),
            0,
          );
        if (score < bestScore) {
          best = option;
          bestScore = score;
        }
      }
      placed[i] = best;
    });
  }
  return placed.map(({ x, y, width, height }) => ({ x, y, width, height }));
}

async function chartPNG(svg, doc = svg.ownerDocument) {
  if (doc.fonts) await doc.fonts.ready;
  const win = doc.defaultView,
    box = svg.viewBox.baseVal;
  const size = pngDimensions(box.width, box.height);
  const copy = svg.cloneNode(true);
  copy.setAttribute("xmlns", "http://www.w3.org/2000/svg");
  copy.setAttribute("width", box.width);
  copy.setAttribute("height", box.height);
  copy.removeAttribute("class");
  copy
    .querySelectorAll("[tabindex]")
    .forEach((node) => node.removeAttribute("tabindex"));
  const source = new win.XMLSerializer().serializeToString(copy);
  const url = win.URL.createObjectURL(
    new win.Blob([source], { type: "image/svg+xml;charset=utf-8" }),
  );
  try {
    const img = new win.Image();
    await new Promise((resolve, reject) => {
      const timeout = win.setTimeout(
        () => reject(new Error("Timed out rendering the chart image.")),
        15000,
      );
      img.onload = () => {
        win.clearTimeout(timeout);
        resolve();
      };
      img.onerror = () => {
        win.clearTimeout(timeout);
        reject(new Error("The browser could not render this chart."));
      };
      img.src = url;
    });
    const canvas = doc.createElement("canvas");
    canvas.width = size.width;
    canvas.height = size.height;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("PNG export requires Canvas 2D support.");
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, size.width, size.height);
    ctx.drawImage(img, 0, 0, size.width, size.height);
    return await new Promise((resolve, reject) =>
      canvas.toBlob(
        (blob) =>
          blob
            ? resolve(blob)
            : reject(
                new Error("PNG encoding failed. Try a smaller selection."),
              ),
        "image/png",
      ),
    );
  } finally {
    win.URL.revokeObjectURL(url);
  }
}

function createReportCharts(doc, data, getScope, getFilters) {
  const ink = "#182f38",
    muted = "#52656d",
    grid = "#e2e9e5";
  const colors = { minimal: "#4e6e9b", medium: "#087568", max: "#a95427" };
  const format = (v, digits = 2) =>
    Number(v).toLocaleString("en-US", { maximumFractionDigits: digits });
  const percent = (v) => `${format(v * 100)}%`;
  const money = (v) =>
    `$${format(v, v > 0 && v < 0.01 ? Math.max(5, 2 - Math.floor(Math.log10(v))) : 4)}`;
  const measure = doc.createElement("canvas").getContext("2d");

  function shape(tag, attrs, parent, value) {
    const e = doc.createElementNS("http://www.w3.org/2000/svg", tag);
    if (tag === "text")
      attrs = {
        "font-family": "Inter, Arial, sans-serif",
        "font-kerning": "normal",
        "font-size": 14,
        fill: muted,
        ...attrs,
      };
    for (const [key, v] of Object.entries(attrs)) e.setAttribute(key, v);
    if (value != null) e.textContent = value;
    parent.append(e);
    return e;
  }
  function text(s, x, y, value, attrs = {}) {
    return shape("text", { x, y, ...attrs }, s, value);
  }
  function modelText(s, x, y, model, attrs = {}, suffix = "") {
    const label = text(s, x, y, modelDisplayName(model) + suffix, {
      "font-size": 15,
      "font-weight": 550,
      "letter-spacing": 0.15,
      fill: ink,
      "data-model-name": model,
      ...attrs,
    });
    shape("title", {}, label, model);
    return label;
  }
  function logo(s, model, x, y, size = 26) {
    const source = modelLogo(model);
    if (source)
      shape(
        "image",
        {
          x,
          y,
          width: size,
          height: size,
          href: source,
          "aria-hidden": "true",
          "data-model-logo": model,
        },
        s,
      );
  }
  function wrap(value, width, size = 14) {
    if (measure) measure.font = `${size}px Inter, Arial, sans-serif`;
    const fits = (s) =>
      (measure ? measure.measureText(s).width : s.length * size * 0.6) <= width;
    const lines = [];
    let line = "";
    for (const word of String(value).split(/\s+/)) {
      if (line && !fits(`${line} ${word}`)) {
        lines.push(line);
        line = "";
      }
      for (const char of (line ? " " : "") + word) {
        if (line && !fits(line + char)) {
          lines.push(line);
          line = "";
        }
        line += char;
      }
    }
    if (line) lines.push(line);
    return lines;
  }
  function lines(s, x, y, value, width, attrs = {}, step = 18) {
    const all = wrap(value, width, attrs["font-size"] || 14);
    all.forEach((line, i) => text(s, x, y + i * step, line, attrs));
    return all.length;
  }
  function scopeText(override = {}) {
    const scope = { ...getScope(), ...override };
    return [
      scope.dataset || (data.datasets || []).join(" + ") || "All datasets",
      {
        full: "Full exam",
        "text-only": "Text-only",
        "image-only": "Image-only",
      }[scope.cohort] || scope.cohort,
      scope.category || "All subjects",
    ].join("  ·  ");
  }
  function chart(id, title, height, options = {}) {
    const {
      width = 1200,
      subtitle = "",
      footer = "",
      display = false,
      scope = {},
      suffix = "",
    } = options;
    const container = doc.getElementById(id);
    container.replaceChildren();
    const toolbar = doc.createElement("div");
    toolbar.className = "chart-toolbar";
    container.append(toolbar);
    const button = doc.createElement("button");
    button.type = "button";
    button.className = "secondary chart-download";
    button.textContent = "Download PNG";
    button.setAttribute("aria-label", `Download PNG: ${title}`);
    toolbar.append(button);
    const status = doc.createElement("span");
    status.className = "note";
    status.setAttribute("role", "status");
    toolbar.append(status);
    const scroll = doc.createElement("div");
    scroll.className = "chart-scroll";
    scroll.tabIndex = 0;
    scroll.setAttribute(
      "aria-label",
      `${title}; scroll horizontally on small screens`,
    );
    container.append(scroll);
    scroll.style.setProperty("--chart-min-width", `${Math.min(width, 1000)}px`);
    const s = shape(
      "svg",
      {
        viewBox: `0 0 ${width} ${height}`,
        width,
        height,
        class: "report-chart",
        role: "img",
        "aria-label": title,
        "font-family": "Inter, Arial, sans-serif",
        "font-size": 14,
        fill: ink,
      },
      scroll,
    );
    shape("title", {}, s, title);
    const fontCSS = doc.getElementById("chart-fonts")?.textContent;
    if (fontCSS) shape("style", { "data-chart-font": "true" }, s, fontCSS);
    shape("rect", { x: 0, y: 0, width, height, fill: "#fff" }, s);
    text(s, 32, 30, "KONKOUR BENCH / RESULTS", {
      "font-size": 11,
      "letter-spacing": 1.5,
      fill: "#087568",
    });
    text(s, 32, 63, title, {
      "font-size": 26,
      "font-weight": 650,
      "letter-spacing": -0.35,
      fill: ink,
    });
    const selection = { ...getScope(), ...scope },
      filters = display ? getFilters() : "";
    let context = scopeText(selection);
    if (display) {
      if (filters) context += `  ·  Display: ${filters}`;
    }
    lines(s, 32, 88, context, width - 64, { "font-size": 13 }, 17);
    text(s, 32, 127, subtitle, { "font-size": 13 });
    shape(
      "line",
      {
        x1: 32,
        x2: width - 32,
        y1: height - 64,
        y2: height - 64,
        stroke: grid,
      },
      s,
    );
    lines(s, 32, height - 42, footer, width - 64, { "font-size": 12 }, 16);
    const dimensions = pngDimensions(width, height);
    button.title = `Save ${dimensions.width} × ${dimensions.height} pixels, with title, labels, and current selection`;
    button.onclick = async () => {
      button.disabled = true;
      status.textContent = "Preparing PNG…";
      try {
        const blob = await chartPNG(s, doc),
          win = doc.defaultView,
          url = win.URL.createObjectURL(blob);
        const a = doc.createElement("a");
        a.href = url;
        a.download =
          [
            "konkour",
            id,
            selection.cohort,
            selection.dataset ||
              (data.datasets || []).join("+") ||
              "all-datasets",
            selection.category,
            filters,
            suffix,
          ]
            .filter(Boolean)
            .join("-")
            .replace(/[^\p{L}\p{N}._-]+/gu, "-") + ".png";
        doc.body.append(a);
        a.click();
        a.remove();
        win.setTimeout(() => win.URL.revokeObjectURL(url), 1000);
        status.textContent = `Saved ${dimensions.width} × ${dimensions.height} PNG`;
      } catch (err) {
        console.error("Chart PNG export failed", id, err);
        status.textContent = `Export failed: ${err.message}`;
      } finally {
        button.disabled = false;
      }
    };
    return s;
  }
  function empty(id, message) {
    const container = doc.getElementById(id);
    container.replaceChildren();
    const p = doc.createElement("p");
    p.className = "empty";
    p.textContent = message;
    container.append(p);
  }
  function marker(s, x, y, tier, size = 6, active = true) {
    const attrs = {
      fill: active ? colors[tier] || colors.medium : "#fff",
      stroke: colors[tier] || colors.medium,
      "stroke-width": active ? 2 : 1.6,
    };
    if (tier === "medium")
      return shape(
        "rect",
        {
          x: x - size,
          y: y - size,
          width: size * 2,
          height: size * 2,
          rx: 1.5,
          ...attrs,
        },
        s,
      );
    if (tier === "max")
      return shape(
        "path",
        {
          d: `M${x} ${y - size - 1}L${x + size + 1} ${y}L${x} ${y + size + 1}L${x - size - 1} ${y}Z`,
          ...attrs,
        },
        s,
      );
    return shape("circle", { cx: x, cy: y, r: size, ...attrs }, s);
  }
  function legend(s, x, y) {
    ["minimal", "medium", "max"].forEach((tier, i) => {
      marker(s, x + i * 128, y - 4, tier, 5);
      text(s, x + 13 + i * 128, y, tier, { "font-size": 13 });
    });
  }

  function accuracy(rows) {
    if (!rows.length)
      return empty(
        "accuracy-chart",
        "No complete configurations match these display filters.",
      );
    const height = 265 + rows.length * 40;
    const s = chart(
      "accuracy-chart",
      "Accuracy by model and reasoning effort",
      height,
      {
        display: true,
        subtitle:
          "Point: observed accuracy  ·  Whisker: 95% Wilson interval  ·  Equal question weights",
        footer:
          "Intervals assume independent question sampling; they do not measure run-to-run variation. Ties retain equal rank.",
      },
    );
    const x = (v) => 410 + v * 620,
      bottom = 180 + (rows.length - 1) * 40 + 20;
    text(s, 32, 158, "MODEL / REQUESTED EFFORT", {
      "font-size": 11,
      "font-weight": 700,
    });
    text(s, 1075, 158, "SCORE", {
      "font-size": 11,
      "text-anchor": "middle",
      "font-weight": 700,
    });
    text(s, 1170, 158, "CORRECT", {
      "font-size": 11,
      "text-anchor": "end",
      "font-weight": 700,
    });
    for (let i = 0; i <= 4; i++) {
      const v = i / 4;
      shape(
        "line",
        {
          x1: x(v),
          x2: x(v),
          y1: 165,
          y2: bottom,
          stroke: i === 1 ? "#b6c7bf" : grid,
          "stroke-dasharray": i === 1 ? "4 4" : "none",
        },
        s,
      );
      text(s, x(v), 154, percent(v), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    rows.forEach((r, i) => {
      const y = 184 + i * 40;
      logo(s, r.model, 32, y - 16);
      modelText(s, 70, y - 2, r.model);
      text(s, 70, y + 14, `${r.model.split("/")[0]} · ${r.tier}`, {
        "font-size": 11,
      });
      if (r.accuracy_interval) {
        const lo = x(r.accuracy_interval.low),
          hi = x(r.accuracy_interval.high),
          color = colors[r.tier];
        shape(
          "line",
          {
            x1: lo,
            x2: hi,
            y1: y,
            y2: y,
            stroke: color,
            "stroke-width": 2.5,
            opacity: 0.7,
          },
          s,
        );
        for (const px of [lo, hi])
          shape(
            "line",
            {
              x1: px,
              x2: px,
              y1: y - 4,
              y2: y + 4,
              stroke: color,
              "stroke-width": 1.5,
            },
            s,
          );
      }
      marker(s, x(r.accuracy), y, r.tier, 5);
      text(s, 1075, y + 5, percent(r.accuracy), {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, `${r.correct}/${r.total}`, {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    text(
      s,
      410,
      bottom + 28,
      "Dashed guide: 25% uniform-guessing expectation",
      { "font-size": 12 },
    );
  }

  function scatter(rows, metric) {
    const metricKey =
        metric === "cost" ? "cost_per_question_usd" : "tokens_per_question",
      frontierKey = metric === "cost" ? "cost_frontier" : "token_frontier";
    const points = rows.filter((r) => r[metricKey] > 0);
    if (!points.length)
      return empty(
        `${metric}-chart`,
        "No complete configurations with positive known accounting.",
      );
    const frontier = points
      .filter((r) => r[frontierKey])
      .sort(
        (a, b) =>
          b.accuracy - a.accuracy ||
          b[metricKey] - a[metricKey] ||
          a.model.localeCompare(b.model),
      );
    const height = Math.max(820, 400 + frontier.length * 70),
      left = 96,
      right = 1148,
      top = 220,
      bottom = height - 190;
    const s = chart(
      `${metric}-chart`,
      metric === "cost" ? "Accuracy and API cost" : "Accuracy and token use",
      height,
      {
        display: true,
        subtitle: `${points.length} complete configurations  ·  Colored points: Pareto frontier  ·  Gray points: other configurations`,
        footer:
          metric === "cost"
            ? "Actual API charges only; subscription usage and unknown charges are excluded. Frontier membership uses all configurations in this scope."
            : "Native tokenizers differ between models. Reasoning is included in completion tokens. Frontier membership uses all configurations in this scope.",
      },
    );
    const low = Math.floor(
        Math.log10(Math.min(...points.map((r) => r[metricKey]))),
      ),
      high = Math.max(
        low + 1,
        Math.ceil(Math.log10(Math.max(...points.map((r) => r[metricKey])))),
      );
    const yMin = Math.max(
      0,
      Math.floor(Math.min(...points.map((r) => r.accuracy)) * 10 - 0.5) / 10,
    );
    const x = (v) =>
        left + ((Math.log10(v) - low) / (high - low)) * (right - left),
      y = (v) => bottom - ((v - yMin) / (1 - yMin)) * (bottom - top);
    text(s, left, 179, "ACCURACY ↑", {
      "font-size": 11,
      "font-weight": 650,
      "letter-spacing": 0.8,
    });
    const yStep = 1 - yMin > 0.5 ? 0.2 : 0.1;
    const yTicks = [yMin, 1];
    for (
      let v = Math.ceil((yMin + 1e-8) / yStep) * yStep;
      v < 1 - 1e-8;
      v += yStep
    )
      yTicks.push(v);
    for (const v of yTicks) {
      shape(
        "line",
        { x1: left, x2: right, y1: y(v), y2: y(v), stroke: "#eaf0ed" },
        s,
      );
      text(s, left - 18, y(v) + 5, percent(v), {
        "text-anchor": "end",
        "font-size": 13,
      });
    }
    for (let i = low; i <= high; i++) {
      shape(
        "line",
        {
          x1: x(10 ** i),
          x2: x(10 ** i),
          y1: bottom,
          y2: bottom + 6,
          stroke: "#b9c8c1",
        },
        s,
      );
      text(
        s,
        x(10 ** i),
        bottom + 28,
        metric === "cost" ? money(10 ** i) : format(10 ** i, 0),
        { "font-size": 13, "text-anchor": "middle" },
      );
    }
    text(
      s,
      (left + right) / 2,
      bottom + 59,
      metric === "cost"
        ? "USD per question · logarithmic scale"
        : "Native tokens per question · logarithmic scale",
      { "text-anchor": "middle", "font-size": 14 },
    );
    legend(s, left + 5, bottom + 102);
    text(
      s,
      right,
      bottom + 102,
      metric === "cost"
        ? "Lower cost and higher accuracy are better"
        : "Fewer tokens and higher accuracy are better",
      { "text-anchor": "end", "font-size": 12 },
    );

    if (frontier.length > 1)
      shape(
        "polyline",
        {
          points: [...frontier]
            .reverse()
            .map((r) => `${x(r[metricKey])},${y(r.accuracy)}`)
            .join(" "),
          fill: "none",
          stroke: "#147d6c",
          "stroke-width": 2.75,
          "stroke-linejoin": "round",
          "data-frontier-line": "true",
        },
        s,
      );
    // Draw frontier markers last so coincident comparison points cannot hide them.
    for (const r of [...points.filter((r) => !r[frontierKey]), ...frontier]) {
      const title = `${modelDisplayName(r.model)} / ${r.tier}: ${percent(r.accuracy)} · ${metric === "cost" ? money(r[metricKey]) : format(r[metricKey], 0)} per question${r[frontierKey] ? " · Pareto frontier" : ""} (${r.model})`;
      const dot = marker(
        s,
        x(r[metricKey]),
        y(r.accuracy),
        r.tier,
        r[frontierKey] ? 8 : 5,
      );
      dot.setAttribute("stroke", r[frontierKey] ? "#ffffff" : "#b4c2bc");
      dot.setAttribute("stroke-width", r[frontierKey] ? "2.5" : "1.5");
      if (!r[frontierKey]) dot.setAttribute("fill", "#f1f5f3");
      dot.setAttribute(
        "data-pareto-point",
        r[frontierKey] ? "frontier" : "comparison",
      );
      dot.setAttribute("tabindex", "0");
      dot.setAttribute("aria-label", title);
      shape("title", {}, dot, title);
    }
    const labels = frontier.map((r) => {
      const name = modelDisplayName(r.model),
        detail = `${r.tier} · ${percent(r.accuracy)}`;
      if (measure) measure.font = "600 16px Inter, Arial, sans-serif";
      const nameWidth = measure
        ? measure.measureText(name).width + name.length * 0.1
        : name.length * 10;
      if (measure) measure.font = "400 12px Inter, Arial, sans-serif";
      const detailWidth = measure
        ? measure.measureText(detail).width
        : detail.length * 8;
      return {
        x: x(r[metricKey]),
        y: y(r.accuracy),
        width: Math.ceil(Math.max(nameWidth, detailWidth) + 32),
        height: 42,
        detail,
      };
    });
    const positions = frontierLabelPositions(
      labels,
      { left, right: right + 10, top: 190, bottom: bottom - 12 },
      points.map((r) => ({ x: x(r[metricKey]), y: y(r.accuracy) })),
      labels.slice(1).map((p, i) => [labels[i], p]),
    );
    frontier.forEach((r, i) => {
      const { x: lx, y: ly } = positions[i];
      const label = shape(
        "g",
        {
          "data-frontier-label": "true",
          "aria-label": `${modelDisplayName(r.model)}, ${labels[i].detail}`,
        },
        s,
      );
      logo(label, r.model, lx, ly + 1, 22);
      const halo = {
        stroke: "#fff",
        "stroke-width": 5,
        "stroke-linejoin": "round",
        "paint-order": "stroke fill",
      };
      modelText(label, lx + 30, ly + 18, r.model, {
        "font-size": 16,
        "font-weight": 600,
        "letter-spacing": 0.1,
        ...halo,
      });
      text(label, lx + 30, ly + 37, labels[i].detail, {
        "font-size": 12,
        ...halo,
      });
    });
    if (!frontier.length)
      text(s, right, 179, "No frontier points match the display filters", {
        "text-anchor": "end",
        "font-size": 12,
      });
  }

  function effort(rows, from, to) {
    if (!rows.length)
      return empty(
        "effort-chart",
        "No completed question pairs for these tiers in this scope.",
      );
    const height = 285 + rows.length * 48,
      bottom = height - 114;
    const s = chart(
      "effort-chart",
      `Reasoning effort: ${from} → ${to}`,
      height,
      {
        subtitle:
          "Net accuracy change on the same questions · Gains minus regressions, in percentage points",
        footer:
          "* Partial pair: only jointly completed questions are compared; missing questions can bias the result. Requested tiers are not equal budgets across models.",
        suffix: `${from}-${to}`,
      },
    );
    const extent =
      Math.ceil(Math.max(0.05, ...rows.map((r) => Math.abs(r.delta))) * 20) /
      20;
    const x = (v) => 750 + (v / extent) * 230;
    text(s, 32, 159, "MODEL / PAIRED QUESTIONS", {
      "font-size": 11,
      "font-weight": 700,
    });
    text(s, 1138, 159, "NET CHANGE", {
      "font-size": 11,
      "font-weight": 700,
      "text-anchor": "end",
    });
    for (const v of [-extent, -extent / 2, 0, extent / 2, extent]) {
      shape(
        "line",
        {
          x1: x(v),
          x2: x(v),
          y1: 175,
          y2: bottom,
          stroke: v === 0 ? "#7f9690" : grid,
          "stroke-width": v === 0 ? 1.5 : 1,
        },
        s,
      );
      text(
        s,
        x(v),
        bottom + 25,
        `${v > 0 ? "+" : ""}${format(v * 100, 1)} pp`,
        { "text-anchor": "middle", "font-size": 12 },
      );
    }
    rows.forEach((r, i) => {
      const y = 196 + i * 48,
        color = r.delta >= 0 ? colors.medium : "#b3433b";
      logo(s, r.model, 32, y - 16);
      modelText(s, 70, y - 2, r.model);
      text(
        s,
        70,
        y + 15,
        `${r.paired}/${r.total} paired${r.paired < r.total ? " *" : ""} · ${r.gained} gained · ${r.lost} lost`,
        { "font-size": 12 },
      );
      shape(
        "line",
        {
          x1: x(0),
          x2: x(r.delta),
          y1: y,
          y2: y,
          stroke: color,
          "stroke-width": 5,
          "stroke-linecap": "round",
        },
        s,
      );
      shape(
        "circle",
        {
          cx: x(r.delta),
          cy: y,
          r: 6,
          fill: color,
          stroke: "#fff",
          "stroke-width": 1.5,
        },
        s,
      );
      text(
        s,
        1138,
        y + 5,
        `${r.delta > 0 ? "+" : ""}${format(r.delta * 100)} pp`,
        {
          "text-anchor": "end",
          "font-size": 16,
          "font-weight": 700,
          fill: color,
        },
      );
    });
  }

  // All bar charts use an explicit zero baseline and label every count.
  function bars(
    s,
    items,
    {
      x = 300,
      top = 190,
      width = 690,
      step = 55,
      max = Math.max(1, ...items.map((d) => d.value)),
      denominator,
      labelX = 32,
    } = {},
  ) {
    const bottom = top + (items.length - 1) * step + 19;
    max = countAxisMaximum(max);
    for (let i = 0; i <= 4; i++) {
      const value = (max * i) / 4,
        px = x + (width * i) / 4;
      shape(
        "line",
        { x1: px, x2: px, y1: top - 28, y2: bottom, stroke: grid },
        s,
      );
      text(s, px, bottom + 24, format(value, 1), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    items.forEach((item, i) => {
      const y = top + i * step;
      text(s, labelX, y + 5, item.label, { "font-size": 14, fill: ink });
      shape(
        "rect",
        {
          x,
          y: y - 13,
          width: (width * item.value) / max,
          height: 27,
          rx: 3,
          fill: item.color || colors.medium,
        },
        s,
      );
      text(
        s,
        x + width + 28,
        y + 5,
        `${format(item.value, 0)}${denominator ? ` / ${format(denominator, 0)}` : ""}`,
        { "font-size": 15, "font-weight": 700, fill: ink },
      );
    });
  }

  function difficulty(bins, panel) {
    if (!bins.some(Boolean))
      return empty(
        "difficulty-chart",
        "No complete configuration panel is available.",
      );
    const total = bins.reduce((a, b) => a + b, 0);
    const s = chart(
      "difficulty-chart",
      "Question difficulty across the complete panel",
      515,
      {
        subtitle: `${total} questions · ${panel} complete configurations · Grouped by the fraction of panel configurations answering correctly`,
        footer:
          "Each question appears once. Configurations from the same model are correlated; these are descriptive votes, not independent observations.",
      },
    );
    bars(
      s,
      bins.map((value, i) => ({
        value,
        label: [
          "Below 25% correct",
          "25% to below 50%",
          "50% to below 75%",
          "75% to 100%",
        ][i],
        color: ["#a15a3b", "#c08b5d", "#6c9e8c", "#087568"][i],
      })),
      {
        top: 200,
        step: 57,
        max: Math.max(4, Math.ceil(Math.max(...bins) / 4) * 4),
      },
    );
    text(s, 645, 428, "Number of questions", {
      "text-anchor": "middle",
      "font-size": 13,
    });
  }

  function coverage(c) {
    const names = {
      complete: "Completed",
      unsupported: "Unsupported",
      skipped: "Image skips",
      infra_error: "Unresolved",
      not_run: "Not run",
      pending: "Pending",
    };
    const palette = {
      complete: "#087568",
      unsupported: "#a0afa6",
      skipped: "#b58b35",
      infra_error: "#b3433b",
      not_run: "#b9c4ca",
      pending: "#4e6e9b",
    };
    const statuses = Object.entries(c.statuses).filter(([, n]) => n > 0);
    const height = 330 + Math.ceil(statuses.length / 3) * 62;
    const s = chart("coverage", "Coverage of configured answer slots", height, {
      subtitle: `${format(c.complete, 0)} / ${format(c.slots, 0)} completed · ${c.questions} questions · ${c.images} with images · ${c.models} models`,
      footer:
        "One slot is one question × model × requested effort. Completed includes invalid answers, which count as wrong. Other statuses remain unranked.",
    });
    let x = 32;
    statuses.forEach(([status, n]) => {
      const width = (n / (c.slots || 1)) * 1136;
      shape(
        "rect",
        { x, y: 173, width, height: 44, fill: palette[status] || "#7f8d91" },
        s,
      );
      x += width;
    });
    statuses.forEach(([status, n], i) => {
      const x = 32 + (i % 3) * 380,
        y = 258 + Math.floor(i / 3) * 62;
      shape(
        "rect",
        {
          x,
          y: y - 12,
          width: 12,
          height: 12,
          rx: 2,
          fill: palette[status] || "#7f8d91",
        },
        s,
      );
      text(s, x + 22, y, names[status] || status, {
        "font-size": 14,
        fill: ink,
      });
      text(
        s,
        x + 22,
        y + 23,
        `${format(n, 0)} · ${percent(n / (c.slots || 1))}`,
        { "font-size": 17, "font-weight": 700, fill: ink },
      );
    });
  }

  function subjects(overall, categories, categoryRows) {
    if (!categories.length || !overall.length)
      return empty("subjects-chart", "No subject results for this selection.");
    const rows = [...overall].sort(
      (a, b) =>
        (b.accuracy ?? -1) - (a.accuracy ?? -1) ||
        a.model.localeCompare(b.model) ||
        a.tier.localeCompare(b.tier),
    );
    const width = Math.max(1200, 420 + categories.length * 175),
      height = 326 + rows.length * 41;
    const s = chart("subjects-chart", "Accuracy by subject", height, {
      width,
      scope: { category: "" },
      subtitle:
        "Each cell uses its subject denominator · Gray cells show completed / selected; they have no accuracy score",
      footer:
        "Cell color encodes accuracy from 0% (light) to 100% (dark). Rows sort by overall accuracy; incomplete configurations follow. Requested effort appears beside each model.",
    });
    const cellWidth = (width - 420) / categories.length;
    text(s, 32, 177, "MODEL / EFFORT", { "font-size": 11, "font-weight": 700 });
    categories.forEach((category, i) => {
      const x = 398 + i * cellWidth + cellWidth / 2;
      text(s, x, 170, category, {
        "text-anchor": "middle",
        "font-size": 16,
        fill: ink,
      });
      text(
        s,
        x,
        191,
        `${categoryRows.find((r) => r.category === category)?.total || 0} questions`,
        { "text-anchor": "middle", "font-size": 12 },
      );
    });
    const lookup = new Map(
      categoryRows.map((r) => [`${r.model}\0${r.tier}\0${r.category}`, r]),
    );
    rows.forEach((r, i) => {
      const y = 209 + i * 41;
      logo(s, r.model, 32, y + 2, 24);
      modelText(s, 68, y + 16, r.model, { "font-size": 14 });
      text(s, 375, y + 16, r.tier, {
        "text-anchor": "end",
        "font-size": 12,
        fill: colors[r.tier],
      });
      categories.forEach((category, j) => {
        const item = lookup.get(`${r.model}\0${r.tier}\0${category}`),
          value = item?.accuracy,
          x = 398 + j * cellWidth;
        const color =
          value == null ? { fill: "#f1f3f2", text: ink } : heatmapColor(value);
        shape(
          "rect",
          {
            x: x + 3,
            y,
            width: cellWidth - 6,
            height: 34,
            rx: 3,
            fill: color.fill,
          },
          s,
        );
        text(
          s,
          x + cellWidth / 2,
          y + 22,
          item
            ? value == null
              ? `${item.complete}/${item.total} done`
              : `${percent(value)} · ${item.correct}/${item.total}`
            : "—",
          {
            "text-anchor": "middle",
            "font-size": 12,
            fill: color.text,
          },
        );
      });
    });
    const legendY = height - 90;
    text(s, 32, legendY + 4, "ACCURACY", { "font-size": 11 });
    [0, 0.25, 0.5, 0.75, 1].forEach((v, i) => {
      const x = 140 + i * 118;
      shape(
        "rect",
        {
          x,
          y: legendY - 10,
          width: 30,
          height: 16,
          rx: 2,
          fill: heatmapColor(v).fill,
        },
        s,
      );
      text(s, x + 38, legendY + 3, percent(v), { "font-size": 12 });
    });
  }

  function answers(q) {
    if (!q.panel)
      return empty(
        "answer-chart",
        "No complete configuration panel is available for answer frequencies.",
      );
    const s = chart(
      "answer-chart",
      `Question ${q.id.split(":").at(-1)} · answer distribution`,
      495,
      {
        width: 720,
        subtitle: `${q.correct}/${q.panel} panel configurations correct · Supplied key: option ${q.expected}`,
        footer:
          "Counts use the fixed complete panel. Model/effort votes are correlated. The supplied key is not independently verified.",
        suffix: q.id,
      },
    );
    bars(
      s,
      [1, 2, 3, 4, 0].map((i) => ({
        value: q.answers[i],
        label: i ? `Option ${i}${i === q.expected ? " · KEY" : ""}` : "Invalid",
        color:
          i === q.expected ? colors.medium : i === 0 ? "#a15a3b" : "#8faaa0",
      })),
      { x: 170, top: 183, width: 400, step: 43, max: q.panel },
    );
    text(s, 370, 418, "Number of panel configurations", {
      "text-anchor": "middle",
      "font-size": 12,
    });
  }

  function pair(result, a, b) {
    const s = chart("pair-chart", "Shared and complementary answers", 600, {
      subtitle: `${result.n} paired questions · Both configurations completed the selected scope`,
      footer:
        "Either correct is an oracle upper bound that knows the answer key, not the accuracy of a deployable ensemble. Invalid completions count as wrong.",
      suffix: `${a.model}-${a.tier}-vs-${b.model}-${b.tier}`,
    });
    text(s, 32, 168, "A", { "font-size": 15, "font-weight": 600, fill: ink });
    text(s, 32, 194, "B", { "font-size": 15, "font-weight": 600, fill: ink });
    logo(s, a.model, 56, 152, 20);
    logo(s, b.model, 56, 178, 20);
    modelText(s, 86, 168, a.model, {}, ` / ${a.tier}`);
    modelText(s, 86, 194, b.model, {}, ` / ${b.tier}`);
    bars(
      s,
      [
        {
          label: "Both correct",
          value: result.both.length,
          color: colors.medium,
        },
        { label: "Only A correct", value: result.a.length, color: "#4e6e9b" },
        { label: "Only B correct", value: result.b.length, color: "#a17a39" },
        { label: "Both wrong", value: result.neither.length, color: "#a15a3b" },
      ],
      { top: 249, step: 58, max: Math.max(1, result.n), denominator: result.n },
    );
    text(s, 645, 495, "Number of paired questions", {
      "text-anchor": "middle",
      "font-size": 13,
    });
  }

  function configurationLabel(s, r, y) {
    logo(s, r.model, 32, y - 16);
    modelText(s, 70, y - 2, r.model);
    text(s, 70, y + 14, `${r.model.split("/")[0]} · ${r.tier}`, {
      "font-size": 11,
    });
  }
  function columnHeading(s, x, y, value, anchor = "start") {
    text(s, x, y, value, {
      "font-size": 11,
      "font-weight": 700,
      "text-anchor": anchor,
    });
  }

  function ranks(rows, total, resamples) {
    if (!rows.length)
      return empty(
        "rank-chart",
        "Rank intervals need at least two complete configurations in this scope.",
      );
    const height = 265 + rows.length * 40,
      bottom = 200 + (rows.length - 1) * 40;
    const s = chart(
      "rank-chart",
      "Rank uncertainty under question resampling",
      height,
      {
        display: true,
        subtitle: `Bar: 5th–95th percentile rank across ${format(resamples, 0)} paired question resamples  ·  Marker: observed rank`,
        footer:
          "Ties share the better rank. Resampling treats questions as exchangeable; it does not measure run-to-run variation. Ranks always use every complete configuration in the scope, including any hidden by display filters.",
      },
    );
    const x = (rank) => 410 + ((rank - 1) / Math.max(1, total - 1)) * 590;
    columnHeading(s, 32, 158, "MODEL / REQUESTED EFFORT");
    columnHeading(s, 1075, 158, "90% RANKS", "middle");
    columnHeading(s, 1170, 158, "FIRST", "end");
    const ticks = new Set([1, total]);
    for (let t = 5; t < total; t += 5) ticks.add(t);
    for (const t of ticks) {
      shape(
        "line",
        { x1: x(t), x2: x(t), y1: 165, y2: bottom, stroke: grid },
        s,
      );
      text(s, x(t), 154, String(t), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    rows.forEach((r, i) => {
      const y = 184 + i * 40,
        color = colors[r.tier] || colors.medium,
        span = r.low === r.high ? String(r.low) : `${r.low}–${r.high}`;
      configurationLabel(s, r, y);
      const bar = shape(
        "line",
        {
          x1: x(r.low),
          x2: x(r.high),
          y1: y,
          y2: y,
          stroke: color,
          "stroke-width": 8,
          "stroke-linecap": "round",
          opacity: 0.3,
        },
        s,
      );
      const title = `${modelDisplayName(r.model)} / ${r.tier}: observed rank ${r.rank}, 90% interval ${span}, median ${r.median}, first in ${percent(r.first)} of resamples`;
      shape("title", {}, bar, title);
      shape("title", {}, marker(s, x(r.rank), y, r.tier, 5), title);
      text(s, 1075, y + 5, span, {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, percent(r.first), {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    text(
      s,
      410,
      bottom + 28,
      "Rank 1 is the most accurate. FIRST: share of resamples ranked first, ties included.",
      { "font-size": 12 },
    );
  }

  function estimates(rows) {
    if (!rows.length)
      return empty(
        "estimate-chart",
        "No unranked configuration in this scope is missing only unresolved answers.",
      );
    const height = 300 + rows.length * 48,
      bottom = 204 + (rows.length - 1) * 48;
    const s = chart(
      "estimate-chart",
      "Unresolved answers: bounds and a difficulty-based estimate",
      height,
      {
        display: true,
        subtitle:
          "Band: every unresolved answer wrong → every one correct  ·  Open marker: rate on completed answers  ·  Filled marker: Rasch estimate",
        footer:
          "The estimate assumes unresolved questions behave like completed questions of the same panel difficulty. It is not a score and never enters rankings. When the held-out check predicts more than the panel actually scored, treat it as optimistic.",
      },
    );
    const values = rows.flatMap((r) => [
      r.correct / r.total,
      r.correct / r.complete,
      r.estimate,
    ]);
    const lo = Math.max(0, Math.floor(Math.min(...values) * 10) / 10),
      x = (v) => 410 + ((v - lo) / (1 - lo) || 0) * 580,
      step = 1 - lo > 0.3 ? 0.1 : 0.05;
    columnHeading(s, 32, 158, "MODEL / REQUESTED EFFORT");
    columnHeading(s, 1050, 158, "ESTIMATE", "middle");
    columnHeading(s, 1170, 158, "UNRESOLVED", "end");
    for (let v = lo; v <= 1 + 1e-9; v += step) {
      shape(
        "line",
        { x1: x(v), x2: x(v), y1: 165, y2: bottom, stroke: grid },
        s,
      );
      text(s, x(v), 154, percent(v), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    rows.forEach((r, i) => {
      const y = 184 + i * 48,
        low = r.correct / r.total,
        high = (r.correct + r.unresolved) / r.total,
        observed = r.correct / r.complete;
      configurationLabel(s, r, y);
      const band = shape(
        "rect",
        {
          x: x(low),
          y: y - 8,
          width: Math.max(2, x(high) - x(low)),
          height: 16,
          rx: 3,
          fill: "#dbe6e0",
        },
        s,
      );
      shape(
        "title",
        {},
        band,
        `Bounds ${percent(low)}–${percent(high)} (${r.unresolved} unresolved of ${r.total})`,
      );
      shape(
        "title",
        {},
        marker(s, x(observed), y, r.tier, 6, false),
        `Completed answers: ${r.correct}/${r.complete} correct (${percent(observed)})`,
      );
      shape(
        "title",
        {},
        marker(s, x(r.estimate), y, r.tier, 6),
        `Rasch estimate ${percent(r.estimate)} · about ${format(r.expected_correct, 1)} ± ${format(r.expected_correct_sd, 1)} of ${r.total} correct`,
      );
      text(s, 1050, y + 5, percent(r.estimate), {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, `${r.unresolved}/${r.total}`, {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    const legendY = bottom + 34;
    shape(
      "rect",
      {
        x: 410,
        y: legendY - 11,
        width: 26,
        height: 14,
        rx: 3,
        fill: "#dbe6e0",
      },
      s,
    );
    text(s, 444, legendY, "Completion bounds", { "font-size": 13 });
    // Shape and color encode requested effort, as in the other charts.
    shape(
      "circle",
      {
        cx: 616,
        cy: legendY - 4,
        r: 5,
        fill: "#fff",
        stroke: ink,
        "stroke-width": 1.6,
      },
      s,
    );
    text(s, 629, legendY, "Completed-answer rate", { "font-size": 13 });
    shape("circle", { cx: 826, cy: legendY - 4, r: 5, fill: ink }, s);
    text(s, 839, legendY, "Rasch estimate", { "font-size": 13 });
    text(s, 1170, legendY, "Marker shape: requested effort", {
      "font-size": 12,
      "text-anchor": "end",
    });
  }

  function separation(groups, panel) {
    if (!groups.length)
      return empty(
        "separation-chart",
        "Variance shares need at least two complete configurations in this scope.",
      );
    const rowHeight = 46,
      top = 200,
      height = top + groups.length * rowHeight + 150,
      stripLeft = 540,
      stripRight = 900,
      sx = (v) => stripLeft + v * (stripRight - stripLeft),
      shareLeft = 960,
      shareWidth = 140,
      bottom = top + groups.length * rowHeight;
    const s = chart(
      "separation-chart",
      "Where complete configurations separate",
      height,
      {
        subtitle:
          "Dots: each complete configuration's accuracy on the group  ·  Bar: share of between-configuration variance in total correct",
        footer:
          "Share = cov(group score, total score) / var(total score) across complete configurations. Subject shares sum to 100%, and a subject's section shares sum to its share; a group that runs against the total is negative. Same-model configurations are correlated.",
      },
    );
    columnHeading(s, 32, 162, "SUBJECT / SECTION");
    columnHeading(s, 420, 162, "QUESTIONS", "middle");
    columnHeading(s, 495, 162, "ALL RIGHT", "middle");
    columnHeading(s, stripLeft, 162, "ACCURACY BY CONFIGURATION");
    columnHeading(s, shareLeft, 162, "SHARE OF VARIANCE");
    for (const v of [0, 0.25, 0.5, 0.75, 1]) {
      shape(
        "line",
        { x1: sx(v), x2: sx(v), y1: top - 12, y2: bottom, stroke: grid },
        s,
      );
      text(s, sx(v), top - 18, percent(v), {
        "font-size": 11,
        "text-anchor": "middle",
      });
    }
    groups.forEach((g, i) => {
      const y = top + i * rowHeight + 22,
        subject = g.level === "subject";
      if (subject && i > 0)
        shape(
          "line",
          {
            x1: 32,
            x2: 1168,
            y1: y - 26,
            y2: y - 26,
            stroke: "#cdd8d3",
          },
          s,
        );
      text(
        s,
        subject ? 32 : 58,
        y + 5,
        subject ? g.category : g.section || "Untagged questions",
        {
          "font-size": subject ? 15 : 14,
          "font-weight": subject ? 650 : 450,
          fill: ink,
        },
      );
      text(s, 420, y + 5, String(g.questions), {
        "font-size": 14,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 495, y + 5, String(g.ceiling), {
        "font-size": 14,
        "text-anchor": "middle",
      });
      shape(
        "line",
        {
          x1: sx(g.min),
          x2: sx(g.max),
          y1: y,
          y2: y,
          stroke: "#b9c8c1",
          "stroke-width": 2,
        },
        s,
      );
      g.correct.forEach((n, j) => {
        const cfg = panel[j],
          v = n / g.questions,
          // A fixed vertical offset keeps coincident configurations visible.
          dot = shape(
            "circle",
            {
              cx: sx(v),
              cy: y + (((j * 7) % 5) - 2) * 3.5,
              r: 4.5,
              fill: colors[cfg.tier] || colors.medium,
              "fill-opacity": 0.7,
              stroke: "#fff",
              "stroke-width": 1,
            },
            s,
          );
        shape(
          "title",
          {},
          dot,
          `${modelDisplayName(cfg.model)} / ${cfg.tier}: ${n}/${g.questions} (${percent(v)})`,
        );
      });
      shape(
        "line",
        {
          x1: sx(g.mean),
          x2: sx(g.mean),
          y1: y - 13,
          y2: y + 13,
          stroke: ink,
          "stroke-width": 2,
        },
        s,
      );
      if (g.variance_share == null) {
        text(s, 1170, y + 5, "—", { "text-anchor": "end" });
        return;
      }
      shape(
        "rect",
        {
          x: shareLeft,
          y: y - 9,
          width: Math.max(0, Math.min(1, g.variance_share)) * shareWidth,
          height: 18,
          rx: 3,
          fill: subject ? colors.medium : "#6c9e8c",
        },
        s,
      );
      text(s, 1170, y + 5, percent(g.variance_share), {
        "font-size": 14,
        "font-weight": 700,
        "text-anchor": "end",
        fill: ink,
      });
    });
    const legendY = bottom + 30;
    ["minimal", "medium", "max"].forEach((tier, i) => {
      shape(
        "circle",
        {
          cx: stripLeft + i * 110,
          cy: legendY - 4,
          r: 4.5,
          fill: colors[tier],
          "fill-opacity": 0.7,
        },
        s,
      );
      text(s, stripLeft + 10 + i * 110, legendY, tier, { "font-size": 12 });
    });
    shape(
      "line",
      {
        x1: stripLeft + 340,
        x2: stripLeft + 340,
        y1: legendY - 12,
        y2: legendY + 4,
        stroke: ink,
        "stroke-width": 2,
      },
      s,
    );
    text(s, stripLeft + 350, legendY, "mean", { "font-size": 12 });
    text(
      s,
      32,
      legendY,
      "ALL RIGHT: questions every configuration answered correctly",
      {
        "font-size": 12,
      },
    );
  }

  function answerLength(rows, summary) {
    const points = rows.filter(
      (r) => r.correct_median_tokens > 0 || r.wrong_median_tokens > 0,
    );
    if (!points.length)
      return empty(
        "length-chart",
        "No complete configurations with known output tokens match these display filters.",
      );
    const sorted = [...points].sort(
      (a, b) =>
        (b.ratio ?? -1) - (a.ratio ?? -1) ||
        a.model.localeCompare(b.model) ||
        a.tier.localeCompare(b.tier),
    );
    const height = 300 + sorted.length * 40,
      bottom = 200 + (sorted.length - 1) * 40;
    const s = chart(
      "length-chart",
      "Output tokens of correct and wrong answers",
      height,
      {
        display: true,
        subtitle:
          "Median output tokens per answer within each complete configuration  ·  Logarithmic scale",
        footer: `${summary.ratio == null ? "" : `Holding question and configuration fixed, wrong answers used ${format(summary.ratio)}× the output tokens of correct answers (${format(summary.answers, 0)} answers). `}Medians also reflect question difficulty: hard questions are both longer and more often wrong. Invalid answers count as wrong.`,
      },
    );
    const values = sorted
      .flatMap((r) => [r.correct_median_tokens, r.wrong_median_tokens])
      .filter((v) => v > 0);
    const low = Math.floor(Math.log10(Math.min(...values))),
      high = Math.max(low + 1, Math.ceil(Math.log10(Math.max(...values)))),
      x = (v) => 410 + ((Math.log10(v) - low) / (high - low)) * 570;
    columnHeading(s, 32, 158, "MODEL / REQUESTED EFFORT");
    columnHeading(s, 1085, 158, "WRONG ÷ CORRECT", "middle");
    columnHeading(s, 1170, 158, "N", "end");
    for (let i = low; i <= high; i++) {
      shape(
        "line",
        { x1: x(10 ** i), x2: x(10 ** i), y1: 165, y2: bottom, stroke: grid },
        s,
      );
      text(s, x(10 ** i), 154, format(10 ** i, 0), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    const right = "#087568",
      wrong = "#a15a3b";
    sorted.forEach((r, i) => {
      const y = 184 + i * 40;
      configurationLabel(s, r, y);
      if (r.correct_median_tokens > 0 && r.wrong_median_tokens > 0)
        shape(
          "line",
          {
            x1: x(r.correct_median_tokens),
            x2: x(r.wrong_median_tokens),
            y1: y,
            y2: y,
            stroke: "#c9d4cf",
            "stroke-width": 3,
          },
          s,
        );
      for (const [value, color, name, n] of [
        [r.correct_median_tokens, right, "correct", r.correct],
        [r.wrong_median_tokens, wrong, "wrong", r.wrong],
      ]) {
        if (!(value > 0)) continue;
        const dot = shape(
          "circle",
          {
            cx: x(value),
            cy: y,
            r: 6,
            fill: color,
            stroke: "#fff",
            "stroke-width": 1.5,
          },
          s,
        );
        shape(
          "title",
          {},
          dot,
          `${modelDisplayName(r.model)} / ${r.tier}: median ${format(value, 0)} output tokens on ${n} ${name} answers`,
        );
      }
      text(s, 1085, y + 5, r.ratio == null ? "—" : `${format(r.ratio, 1)}×`, {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, `${r.correct} / ${r.wrong}`, {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    const legendY = bottom + 34;
    for (const [i, color, label] of [
      [0, right, "Correct answers (median)"],
      [1, wrong, "Wrong answers (median)"],
    ]) {
      shape(
        "circle",
        { cx: 416 + i * 220, cy: legendY - 4, r: 6, fill: color },
        s,
      );
      text(s, 428 + i * 220, legendY, label, { "font-size": 13 });
    }
    text(s, 1170, legendY, "N: correct / wrong answers", {
      "font-size": 12,
      "text-anchor": "end",
    });
  }

  return {
    accuracy,
    scatter,
    effort,
    difficulty,
    coverage,
    subjects,
    answers,
    pair,
    ranks,
    estimates,
    separation,
    answerLength,
  };
}

if (typeof module !== "undefined" && module.exports)
  module.exports = {
    countAxisMaximum,
    heatmapColor,
    pngDimensions,
    frontierLabelPositions,
    chartPNG,
    createReportCharts,
  };

"use strict";

// Pure helpers are shared with the Node regression tests.
function matchesScope(o, s) {
  return (
    (!s.dataset || o.dataset === s.dataset) &&
    (!s.category || o.category === s.category) &&
    (s.cohort === "full" ||
      (s.cohort === "image-only" ? o.has_image : !o.has_image))
  );
}
function difficultyBins(questions) {
  const bins = [0, 0, 0, 0];
  for (const q of questions) {
    if (q.panel > 0) bins[Math.min(3, Math.floor((q.correct / q.panel) * 4))]++;
  }
  return bins;
}
function compareOutcomes(a, b) {
  const other = new Map(
    b.filter((o) => o.status === "complete").map((o) => [o.question, o]),
  );
  const result = { both: [], a: [], b: [], neither: [], agreement: 0, n: 0 };
  for (const o of a) {
    const p = other.get(o.question);
    if (o.status !== "complete" || !p) continue;
    result.n++;
    if (
      !o.invalid &&
      !p.invalid &&
      o.answer >= 1 &&
      o.answer <= 4 &&
      o.answer === p.answer
    )
      result.agreement++;
    result[
      o.correct ? (p.correct ? "both" : "a") : p.correct ? "b" : "neither"
    ].push(o.question);
  }
  return result;
}
function rankSpan(r) {
  return r.low === r.high ? String(r.low) : `${r.low}–${r.high}`;
}
// Name the configurations at a separation group's lowest or highest score.
function separationExtreme(group, panel, pick) {
  if (!group.correct.length) return null;
  const value = pick(...group.correct);
  return {
    accuracy: value / group.questions,
    configurations: panel.filter((_, i) => group.correct[i] === value),
  };
}
function sortedRows(rows, key, ascending) {
  return [...rows].sort((a, b) => {
    if (a[key] == null) return b[key] == null ? 0 : 1;
    if (b[key] == null) return -1;
    const diff =
      typeof a[key] === "string"
        ? a[key].localeCompare(b[key])
        : Number(a[key]) - Number(b[key]);
    return (
      (ascending ? diff : -diff) ||
      a.model.localeCompare(b.model) ||
      a.tier.localeCompare(b.tier)
    );
  });
}

function initReport(doc, data) {
  data = { ...data, rows: data.rows || [], outcomes: data.outcomes || [] };
  const $ = (id) => doc.getElementById(id);
  const number = (v, digits = 1) =>
    v == null
      ? "—"
      : Number(v).toLocaleString("en-US", { maximumFractionDigits: digits });
  const pct = (v) => (v == null ? "—" : `${number(v * 100, 2)}%`);
  const usd = (v) => (v == null ? "—" : `$${number(v, v < 0.01 ? 5 : 4)}`);
  const signed = (v) =>
    v == null ? "—" : `${v > 0 ? "+" : ""}${number(v * 100, 2)} pp`;
  // Compact one-decimal forms keep the wider diagnostic tables readable.
  const pct1 = (v) => (v == null ? "—" : `${number(v * 100, 1)}%`);
  const points = (v) =>
    v == null
      ? "—"
      : `${v > 0 ? "+" : v < 0 ? "−" : ""}${number(Math.abs(v) * 100, 1)} pp`;
  const short = modelDisplayName;
  const key = (r) => `${r.model}\u0000${r.tier}`;
  const label = (r) => `${short(r.model)} / ${r.tier}`;
  const qNumber = (id) => id.slice(id.lastIndexOf(":") + 1);
  const bank = new Map((data.question_bank || []).map((q) => [q.id, q]));
  let view,
    scope,
    scopeRows = [],
    scopedOutcomes = [],
    selectedQuestion = "",
    sortKey = "accuracy",
    ascending = false;

  const charts = createReportCharts(
    doc,
    data,
    () => scope,
    () =>
      [$("model-search").value.trim(), $("tier").value]
        .filter(Boolean)
        .join(" · "),
  );

  function el(tag, text, parent, className) {
    const node = doc.createElement(tag);
    if (text != null) node.textContent = text;
    if (className) node.className = className;
    if (parent) parent.append(node);
    return node;
  }
  function clear(id) {
    const node = $(id);
    node.replaceChildren();
    return node;
  }
  function empty(parent, text = "No results for this selection.") {
    return el("p", text, parent, "empty");
  }
  function table(id, columns, rows, sortable = false) {
    const wrap = clear(id);
    if (!rows.length) {
      empty(wrap);
      return;
    }
    wrap.tabIndex = 0;
    wrap.setAttribute(
      "aria-label",
      `${id.replaceAll("-", " ")} table; scroll horizontally if needed`,
    );
    const t = el("table", null, wrap),
      tr = el("tr", null, el("thead", null, t));
    for (const c of columns) {
      const th = el("th", null, tr);
      th.scope = "col";
      if (sortable && c.key) {
        th.setAttribute(
          "aria-sort",
          sortKey === c.key ? (ascending ? "ascending" : "descending") : "none",
        );
        const b = el(
          "button",
          c.label + (sortKey === c.key ? (ascending ? " ↑" : " ↓") : ""),
          th,
        );
        b.type = "button";
        b.onclick = () => {
          ascending = sortKey === c.key ? !ascending : c.key === "model";
          sortKey = c.key;
          renderPerformance();
        };
      } else th.textContent = c.label;
    }
    const body = el("tbody", null, t);
    for (const row of rows) {
      const tr = el("tr", null, body);
      for (const c of columns) {
        const td = el("td", null, tr, c.className),
          content = c.format(row);
        if (content?.nodeType) td.append(content);
        else td.textContent = content ?? "—";
      }
    }
  }
  function modelIdentity(model, detail = "") {
    const node = el("span", null, null, "model-identity"),
      source = modelLogo(model);
    node.title = model;
    if (source) {
      const img = el("img", null, node, "model-logo");
      img.src = source;
      img.alt = "";
      img.width = 24;
      img.height = 24;
      img.setAttribute("aria-hidden", "true");
    }
    const copy = el("span", null, node, "model-copy");
    el("span", short(model), copy, "model-name");
    if (detail) el("span", detail, copy, "sub");
    return node;
  }
  function modelCell(r) {
    return modelIdentity(
      r.model,
      `${r.model.split("/")[0]} · ${r.tier} → ${r.sent_efforts || "—"}`,
    );
  }
  function qLink(id, parent) {
    const b = el("button", `Q${qNumber(id)}`, parent, "question-link");
    b.type = "button";
    b.title = id;
    b.onclick = () => {
      selectedQuestion = id;
      $("question-search").value = "";
      $("question-filter").value = "all";
      renderQuestions();
      $("questions").scrollIntoView({ behavior: "instant" });
      $("question-detail").tabIndex = -1;
      $("question-detail").focus({ preventScroll: true });
    };
    return b;
  }
  function filteredRows() {
    const search = $("model-search").value.toLowerCase().trim(),
      tier = $("tier").value;
    return scopeRows.filter(
      (r) =>
        (!search || modelMatchesQuery(r.model, search)) &&
        (!tier || r.tier === tier),
    );
  }
  function renderOverview() {
    const box = clear("insights");
    for (const insight of view.insights) {
      const card = el("article", null, box, "insight");
      el("div", insight.label, card, "insight-label");
      el("div", insight.value, card, "insight-value");
      el("p", insight.detail, card);
    }
    if (!view.insights.length) empty(box);
    charts.coverage(view.coverage);
  }
  function renderPerformance() {
    const rows = filteredRows(),
      ranked = rows.filter((r) => r.accuracy != null),
      visible = new Set(rows.map(key)),
      rankIntervals = new Map(view.ranks.map((r) => [key(r), r]));
    charts.accuracy(sortedRows(ranked, "accuracy", false));
    charts.ranks(
      view.ranks.filter((r) => visible.has(key(r))),
      view.panel.length,
      view.rank_resamples,
    );
    const allRanked = scopeRows.filter((r) => r.accuracy != null);
    table(
      "leaderboard",
      [
        {
          label: "Rank",
          format: (r) =>
            1 + allRanked.filter((v) => v.accuracy > r.accuracy).length,
        },
        {
          key: "model",
          label: "Model / requested → sent",
          format: modelCell,
          className: "model",
        },
        { key: "accuracy", label: "Accuracy", format: (r) => pct(r.accuracy) },
        { label: "Correct", format: (r) => `${r.correct} / ${r.total}` },
        {
          label: "95% interval",
          format: (r) =>
            r.accuracy_interval
              ? `${pct(r.accuracy_interval.low)}–${pct(r.accuracy_interval.high)}`
              : "—",
        },
        {
          label: "Rank, 90% resampled",
          format: (r) => {
            const interval = rankIntervals.get(key(r));
            return interval ? rankSpan(interval) : "—";
          },
        },
        {
          key: "cost_usd",
          label: "API cost / exam",
          format: (r) => usd(r.cost_usd),
        },
        {
          key: "tokens_per_question",
          label: "Tokens / Q",
          format: (r) => number(r.tokens_per_question, 0),
        },
        {
          label: "Frontier",
          format: (r) =>
            [r.cost_frontier && "cost", r.token_frontier && "tokens"]
              .filter(Boolean)
              .join(" · ") || "—",
        },
      ],
      sortedRows(ranked, sortKey, ascending),
      true,
    );
    const estimates = view.estimates.filter((e) => visible.has(key(e)));
    charts.estimates(estimates);
    table(
      "estimate-table",
      [
        {
          label: "Model / effort",
          format: (e) => modelIdentity(e.model, e.tier),
          className: "model",
        },
        {
          label: "Correct / done / total",
          format: (e) => `${e.correct} / ${e.complete} / ${e.total}`,
        },
        {
          label: "Bounds",
          format: (e) =>
            `${pct1(e.correct / e.total)}–${pct1((e.correct + e.unresolved) / e.total)}`,
        },
        {
          label: "Panel accuracy: unresolved vs done",
          format: (e) =>
            `${pct1(e.unresolved_panel_accuracy)} vs ${pct1(e.completed_panel_accuracy)}`,
        },
        {
          label: "Same model, unresolved Qs",
          format: (e) =>
            e.same_model.length
              ? e.same_model
                  .map((c) => `${c.label} ${c.n}/${e.unresolved}`)
                  .join(" · ")
              : "—",
        },
        {
          label: "Estimate (correct ± SD)",
          format: (e) =>
            `${pct1(e.estimate)} (${number(e.expected_correct, 1)} ± ${number(e.expected_correct_sd, 1)})`,
        },
        {
          label: "Held-out: predicted vs actual",
          format: (e) =>
            e.check_predicted == null
              ? "—"
              : `${pct1(e.check_predicted)} vs ${pct1(e.unresolved_panel_accuracy)}`,
        },
      ],
      estimates,
    );
    const unranked = rows.filter((r) => r.accuracy == null);
    $("unranked-summary").textContent =
      `${unranked.length} unranked configurations · coverage and possible completion bounds`;
    table(
      "unranked-table",
      [
        {
          label: "Model / requested → sent",
          format: modelCell,
          className: "model",
        },
        { label: "Status", format: (r) => r.status },
        { label: "Completed", format: (r) => `${r.complete} / ${r.total}` },
        { label: "Known correct", format: (r) => r.correct },
        { label: "Unresolved", format: (r) => r.failed },
        {
          label: "Unsupported / skipped",
          format: (r) => `${r.unsupported} / ${r.skipped}`,
        },
        {
          label: "Completion bounds",
          format: (r) =>
            r.completion_bounds
              ? `${pct(r.completion_bounds.low)}–${pct(r.completion_bounds.high)}`
              : "—",
        },
      ],
      unranked,
    );
    charts.scatter(ranked, "cost");
    charts.scatter(ranked, "tokens");
    const distribution = (r, name, format) => {
      const d = r[name];
      return d?.n ? `${format(d.median)} / ${format(d.p90)} (n=${d.n})` : "—";
    };
    table(
      "resource-table",
      [
        { label: "Model / effort", format: modelCell, className: "model" },
        { label: "Completed", format: (r) => `${r.complete}/${r.total}` },
        {
          label: "USD median / p90",
          format: (r) => distribution(r, "cost_distribution", usd),
        },
        {
          label: "Tokens median / p90",
          format: (r) =>
            distribution(r, "token_distribution", (v) => number(v, 0)),
        },
        {
          label: "Seconds median / p90",
          format: (r) =>
            distribution(r, "latency_seconds", (v) => number(v, 1)),
        },
        { label: "Invalid answers", format: (r) => r.invalid },
      ],
      rows.filter((r) => r.complete > 0),
    );
    const length = view.length;
    charts.answerLength(
      length.rows.filter((r) => visible.has(key(r))),
      length,
    );
    const summary = [];
    if (length.ratio != null)
      summary.push(
        `Holding question and configuration fixed, wrong answers used ${number(length.ratio, 2)}× the output tokens of correct answers (${number(length.answers, 0)} answers; ${length.contrast_questions} questions have both correct and wrong answers).`,
      );
    if (length.longest)
      summary.push(
        `Each complete configuration's longest 10% of answers were ${pct(length.longest_correct / length.longest)} correct (${length.longest_correct} of ${length.longest}), versus ${pct(length.answers_correct / length.answers)} across all ${number(length.answers, 0)} answers with known output tokens.`,
      );
    clear("length-summary").textContent = summary.length
      ? summary.join(" ")
      : "No complete configurations with known output tokens in this scope.";
  }
  function renderEffort() {
    const [from, to] = $("effort-pair").value.split("|");
    const rows = view.efforts
      .filter((e) => e.from === from && e.to === to && e.paired > 0)
      .sort((a, b) => b.delta - a.delta || a.model.localeCompare(b.model));
    charts.effort(rows, from, to);
    table(
      "effort-table",
      [
        {
          label: "Model",
          format: (r) => modelIdentity(r.model),
          className: "model",
        },
        {
          label: "Paired / selected",
          format: (r) =>
            `${r.paired} / ${r.total}${r.paired < r.total ? " · partial*" : ""}`,
        },
        { label: "Gains", format: (r) => `+${r.gained}` },
        { label: "Regressions", format: (r) => `−${r.lost}` },
        { label: "Net accuracy", format: (r) => signed(r.delta) },
        {
          label: "Cost multiplier",
          format: (r) =>
            r.cost_ratio == null ? "—" : `${number(r.cost_ratio, 2)}×`,
        },
        {
          label: "Token multiplier",
          format: (r) =>
            r.token_ratio == null ? "—" : `${number(r.token_ratio, 2)}×`,
        },
      ],
      rows,
    );
  }
  function renderSubjects() {
    const categoryRows = data.rows.filter(
      (r) =>
        r.cohort === scope.cohort && r.dataset === scope.dataset && r.category,
    );
    const categories = [...new Set(categoryRows.map((r) => r.category))];
    const overall = data.rows.filter(
      (r) =>
        r.cohort === scope.cohort && r.dataset === scope.dataset && !r.category,
    );
    charts.subjects(overall, categories, categoryRows);
    const mix = clear("subject-mix");
    for (const category of categories) {
      const n = categoryRows.find((r) => r.category === category)?.total || 0;
      const card = el("div", null, mix, "subject-card");
      el("span", category, card).dir = "auto";
      el("b", `${n} questions`, card);
    }
    table(
      "heatmap",
      [
        { label: "Model / effort", format: modelCell, className: "model" },
        ...categories.map((category) => ({
          label: category,
          format: (r) => {
            const item = categoryRows.find(
                (v) => key(v) === key(r) && v.category === category,
              ),
              cell = el("div", null, null, "heat-cell");
            if (!item) {
              cell.textContent = "—";
              return cell;
            }
            if (item.accuracy != null) {
              cell.textContent = pct(item.accuracy);
              el("span", `${item.correct} / ${item.total}`, cell);
              cell.style.background = `rgba(74, 141, 107, ${0.07 + item.accuracy * 0.23})`;
              cell.style.padding = "6px 12px";
              cell.style.borderRadius = "4px";
            } else {
              cell.textContent = item.status;
              el("span", `${item.complete} / ${item.total} completed`, cell);
            }
            if (category === scope.category)
              cell.classList.add("selected-column");
            return cell;
          },
        })),
      ],
      overall,
    );
  }
  function renderSeparation() {
    charts.separation(view.separation, view.panel);
    const extreme = (g, pick) => {
      const e = separationExtreme(g, view.panel, pick);
      if (!e) return "—";
      const names =
        e.configurations.length > 3
          ? `${e.configurations.length} configurations`
          : e.configurations.map(label).join(", ");
      return `${pct(e.accuracy)} · ${names}`;
    };
    table(
      "separation-table",
      [
        {
          label: "Subject / section",
          format: (g) => {
            const node = el(
              "span",
              g.level === "subject"
                ? g.category
                : `↳ ${g.section || "Untagged questions"}`,
            );
            node.dir = "auto";
            return node;
          },
        },
        { label: "Questions", format: (g) => g.questions },
        {
          label: "All right / none right",
          format: (g) => `${g.ceiling} / ${g.floor}`,
        },
        { label: "Mean accuracy", format: (g) => pct(g.mean) },
        { label: "Lowest", format: (g) => extreme(g, Math.min) },
        { label: "Highest", format: (g) => extreme(g, Math.max) },
        {
          label: "Share of variance",
          format: (g) => pct(g.variance_share),
        },
      ],
      view.separation,
    );
  }
  function renderDifficulty() {
    charts.difficulty(difficultyBins(view.questions), view.coverage.ranked);
  }
  function renderQuestions() {
    const query = $("question-search").value.trim().toLowerCase(),
      filter = $("question-filter").value;
    const questions = view.questions.filter((q) => {
      const content = bank.get(q.id),
        match =
          !query ||
          [q.id, q.category, content?.text].some((t) =>
            t?.toLowerCase().includes(query),
          );
      return (
        match &&
        (filter === "all" ||
          (filter === "review" && q.review) ||
          (filter === "hard" && q.panel > 0 && q.correct * 2 < q.panel) ||
          (filter === "image" && q.has_image))
      );
    });
    $("panel-note").textContent =
      `${view.coverage.ranked} complete configurations in the difficulty panel. All ${view.coverage.configurations} selected configurations remain visible in the answer details. Questions are ordered from lowest to highest panel accuracy.`;
    $("question-count").textContent =
      `${questions.length} of ${view.questions.length} questions shown`;
    if (!questions.some((q) => q.id === selectedQuestion))
      selectedQuestion = questions[0]?.id || "";
    const list = clear("question-list");
    for (const q of questions) {
      const button = el("button", null, list, "question-button");
      button.type = "button";
      button.setAttribute("aria-pressed", String(q.id === selectedQuestion));
      button.title = q.id;
      el(
        "strong",
        `Q${qNumber(q.id)} · ${q.panel ? `${q.correct}/${q.panel} correct` : "No complete panel"}`,
        button,
      );
      el(
        "span",
        `${q.category}${q.section ? ` · ${q.section}` : ""}${q.has_image ? " · image" : ""}${q.review ? " · review" : ""}`,
        button,
        "sub",
      ).dir = "auto";
      const bar = el("div", null, button, "mini-bar");
      el("span", null, bar).style.width =
        `${q.panel ? (q.correct / q.panel) * 100 : 0}%`;
      button.onclick = () => {
        selectedQuestion = q.id;
        renderQuestions();
      };
    }
    if (!questions.length) {
      empty(list, "No questions match these filters.");
      empty(
        clear("question-detail"),
        "Choose another filter to inspect a question.",
      );
      return;
    }
    renderQuestion(view.questions.find((q) => q.id === selectedQuestion));
  }
  function renderQuestion(q) {
    const box = clear("question-detail"),
      content = bank.get(q.id);
    el("h3", `Question ${qNumber(q.id)}`, box);
    el(
      "p",
      `${q.id} · ${q.category}${q.section ? ` · ${q.section}` : ""}`,
      box,
      "note",
    ).dir = "auto";
    if (q.review)
      el(
        "p",
        "Review candidate: most complete configurations chose the same answer other than the supplied key.",
        box,
        "callout",
      );
    if (content) {
      el("div", content.text, box, "question-text").lang = "fa";
      const addImage = (src, parent, alt) => {
        if (!/^data:image\/(png|jpeg|gif);base64,/.test(src || "")) return;
        const img = el("img", null, parent);
        img.src = src;
        img.alt = alt;
      };
      addImage(content.image, box, `Question ${content.number} figure`);
      for (const option of content.options) {
        const row = el(
          "div",
          null,
          box,
          `option${option.number === q.expected ? " key" : ""}`,
        );
        el(
          "span",
          `${option.number}${option.number === q.expected ? " · KEY" : ""}`,
          row,
          "option-label",
        );
        const body = el("div", null, row, "option-content");
        el("div", option.text, body, "option-text").lang = "fa";
        addImage(option.image, body, `Option ${option.number} figure`);
      }
      if (typeof renderMathInElement === "function")
        renderMathInElement(box, {
          delimiters: [
            { left: "$$", right: "$$", display: true },
            { left: "\\[", right: "\\]", display: true },
            { left: "\\(", right: "\\)", display: false },
            { left: "$", right: "$", display: false },
          ],
          throwOnError: false,
          trust: false,
        });
    } else
      el(
        "p",
        "Question text is unavailable in this report snapshot.",
        box,
        "note",
      );
    el("div", null, box).id = "answer-chart";
    charts.answers(q);
    el(
      "p",
      `${q.correct} / ${q.panel} panel configurations correct. Key: option ${q.expected}.`,
      box,
      "note",
    );
    const wrap = el("div", null, box, "table-wrap");
    wrap.id = "question-answers";
    table(
      "question-answers",
      [
        {
          label: "Model / effort",
          format: (o) => modelIdentity(o.model, o.tier),
          className: "model",
        },
        {
          label: "Answer",
          format: (o) =>
            o.status === "complete"
              ? o.invalid
                ? "Invalid"
                : o.answer || "Invalid"
              : "—",
        },
        {
          label: "Result",
          format: (o) =>
            o.status === "complete"
              ? o.correct
                ? "Correct"
                : "Wrong"
              : o.status,
        },
        { label: "USD", format: (o) => usd(o.usage.cost) },
        { label: "Tokens", format: (o) => number(o.usage.total_tokens, 0) },
        {
          label: "Audit detail",
          format: (o) => {
            const d = el("details");
            el("summary", "Inspect", d).title =
              `${o.model} / ${o.tier} audit detail`;
            el(
              "p",
              `Backend: ${o.backend || "—"} · Provider: ${o.provider || "—"} · Sent effort: ${o.sent_effort || "—"}`,
              d,
              "note",
            );
            el(
              "p",
              `Selected run: ${o.run_id || "—"} · Original inference: ${o.origin_run || "—"} · ${o.cached ? "Reused cache" : "Recorded attempt"}`,
              d,
              "note",
            );
            if (o.invalid || o.note)
              el(
                "p",
                [o.invalid, o.note].filter(Boolean).join(" · "),
                d,
                "note",
              );
            el("code", o.request_hash || "No request dispatched", d);
            return d;
          },
        },
      ],
      scopedOutcomes.filter((o) => o.question === q.id),
    );
  }
  function setupComparison() {
    const ranked = scopeRows.filter((r) => r.accuracy != null);
    for (const [id, index] of [
      ["compare-a", 0],
      ["compare-b", 1],
    ]) {
      const previous = $(id).value,
        select = clear(id);
      for (const r of ranked) {
        const option = el("option", label(r), select);
        option.value = key(r);
      }
      const fallback = ranked[Math.min(index, ranked.length - 1)];
      select.value = ranked.some((r) => key(r) === previous)
        ? previous
        : fallback
          ? key(fallback)
          : "";
      select.disabled = !ranked.length;
    }
    renderComparison();
  }
  function renderComparison() {
    const box = clear("pair-results"),
      a = $("compare-a").value,
      b = $("compare-b").value;
    if (!a || !b) {
      empty(box, "No complete configurations available for paired comparison.");
      return;
    }
    const pair = compareOutcomes(
        scopedOutcomes.filter((o) => key(o) === a),
        scopedOutcomes.filter((o) => key(o) === b),
      ),
      chart = el("div", null, box, "chart-card");
    chart.id = "pair-chart";
    charts.pair(
      pair,
      scopeRows.find((r) => key(r) === a),
      scopeRows.find((r) => key(r) === b),
    );
    el(
      "p",
      `Net advantage of B: ${signed(pair.n ? (pair.b.length - pair.a.length) / pair.n : null)} · Same valid answer: ${pair.agreement}/${pair.n} · Either correct (oracle): ${pct(pair.n ? (pair.n - pair.neither.length) / pair.n : null)}`,
      box,
      "callout",
    );
    for (const [title, ids] of [
      ["Both wrong", pair.neither],
      ["Only A correct", pair.a],
      ["Only B correct", pair.b],
    ]) {
      if (!ids.length) continue;
      const block = el("div", null, box);
      block.style.marginTop = "18px";
      el("h3", title, block);
      const links = el("div", null, block, "question-links");
      ids.forEach((id) => qLink(id, links));
    }
  }
  function renderHealth() {
    const box = clear("health"),
      c = view.coverage,
      failures = el("div", null, box, "health-card");
    el("h3", "What prevented a valid answer?", failures);
    if (!c.reasons.length)
      el("p", "All selected outcomes are valid completions.", failures, "note");
    else {
      const list = el("ul", null, failures);
      for (const reason of c.reasons)
        el("li", `${number(reason.n, 0)} × ${reason.label}`, list);
    }
    const accounting = el("div", null, box, "health-card");
    el("h3", "Accounting quality", accounting);
    const completed = scopedOutcomes.filter((o) => o.status === "complete"),
      paid = completed.filter((o) => o.backend !== "claude");
    const facts = [
      `${c.invalid} invalid completions are scored wrong.`,
      `${paid.filter((o) => o.usage.cost == null).length} / ${paid.length} completed paid-API answers have unknown charges.`,
      `${completed.filter((o) => o.backend === "claude").length} subscription completions have no reported actual dollar charge.`,
      `${completed.filter((o) => o.usage.total_tokens == null).length} completed answers have unknown total tokens.`,
      `${completed.filter((o) => o.duration_seconds == null).length} completed answers have unavailable or invalid original timing.`,
      `${c.token_anomalies} completed answers report reasoning tokens greater than completion tokens. Values are preserved as received; do not infer a reasoning share from these records.`,
    ];
    const list = el("ul", null, accounting);
    facts.forEach((f) => el("li", f, list));
    table(
      "providers",
      [
        {
          label: "Model",
          format: (e) => modelIdentity(e.model),
          className: "model",
        },
        { label: "Provider", format: (e) => e.provider },
        {
          label: "Tiers served",
          format: (e) => e.tiers.map((t) => `${t.label} ${t.n}`).join(" · "),
        },
        { label: "Answers", format: (e) => e.answers },
        { label: "Invalid", format: (e) => e.invalid },
        { label: "Accuracy", format: (e) => pct1(e.correct / e.answers) },
        { label: "Expected", format: (e) => pct1(e.expected) },
        { label: "vs expected", format: (e) => points(e.excess) },
        {
          label: "vs other providers (95% CI)",
          format: (e) =>
            e.gap == null
              ? "—"
              : `${points(e.gap)} ${e.gap_interval ? `(${points(e.gap_interval.low)} to ${points(e.gap_interval.high)})` : "· no CI: a side has < 10 answers"}`,
        },
      ],
      view.providers,
    );
  }
  function renderProvenance() {
    table(
      "spending",
      [
        { label: "Run", format: (r) => r.run_id },
        { label: "New attempts", format: (r) => r.attempts },
        {
          label: "Subscription attempts",
          format: (r) => r.subscription_attempts || 0,
        },
        { label: "Known charges", format: (r) => usd(r.known_cost_usd) },
        {
          label: "Included retry cost",
          format: (r) => usd(r.known_retry_cost_usd),
        },
        {
          label: "Unknown paid charges",
          format: (r) =>
            data.backend === "claude" ? "N/A" : r.attempts_with_unknown_cost,
        },
      ],
      data.run_spending || [],
    );
    table(
      "sources",
      [
        {
          label: "Run / started (UTC)",
          format: (r) => {
            const node = el("span", r.id);
            el(
              "span",
              r.created.replace("T", " ").replace(/\.\d+Z$/, " UTC"),
              node,
              "sub",
            );
            return node;
          },
        },
        {
          label: "Backend / CLI version",
          format: (r) =>
            `${r.backend}${r.cli_version ? ` / ${r.cli_version}` : ""}`,
        },
        { label: "Selected jobs", format: (r) => r.selected_jobs },
        { label: "Origin answers", format: (r) => r.origin_answers },
        {
          label: "Timeout / budget",
          format: (r) =>
            `${number(r.timeout_seconds / 60)} min / ${usd(r.budget_usd)}`,
        },
      ],
      data.source_runs || [],
    );
    const meta = clear("dataset-metadata");
    for (const [id, content] of Object.entries(data.dataset_metadata || {})) {
      el("h3", id, meta);
      el("pre", JSON.stringify(content, null, 2), meta);
    }
    if (data.warnings?.length) {
      const list = el("ul", null, meta);
      for (const warning of data.warnings) el("li", warning, list);
    }
  }
  function render() {
    scope = {
      cohort: $("cohort").value,
      dataset: $("dataset").value,
      category: $("category").value,
    };
    scopeRows = data.rows.filter(
      (r) =>
        r.cohort === scope.cohort &&
        r.dataset === scope.dataset &&
        r.category === scope.category,
    );
    scopedOutcomes = data.outcomes.filter((o) => matchesScope(o, scope));
    view = (data.analysis || []).find(
      (v) =>
        v.cohort === scope.cohort &&
        v.dataset === scope.dataset &&
        v.category === scope.category,
    ) || {
      coverage: {
        questions: 0,
        images: 0,
        models: 0,
        configurations: 0,
        ranked: 0,
        slots: 0,
        complete: 0,
        statuses: {},
        reasons: [],
        invalid: 0,
        token_anomalies: 0,
      },
      insights: [],
      efforts: [],
      questions: [],
      panel: [],
      ranks: [],
      rank_resamples: 0,
      separation: [],
      length: { rows: [], answers: 0, answers_correct: 0, longest: 0 },
      estimates: [],
      providers: [],
    };
    $("scope-summary").textContent =
      `${view.coverage.questions} questions · ${view.coverage.ranked}/${view.coverage.configurations} configurations ranked · ${scope.cohort}${scope.category ? ` · ${scope.category}` : ""}`;
    renderOverview();
    renderDifficulty();
    renderPerformance();
    renderEffort();
    renderSubjects();
    renderSeparation();
    renderQuestions();
    setupComparison();
    renderHealth();
  }
  for (const id of ["cohort", "dataset", "category"])
    $(id).addEventListener("change", render);
  $("model-search").addEventListener("input", renderPerformance);
  $("tier").addEventListener("change", renderPerformance);
  $("effort-pair").addEventListener("change", renderEffort);
  $("question-search").addEventListener("input", renderQuestions);
  $("question-filter").addEventListener("change", renderQuestions);
  for (const id of ["compare-a", "compare-b"])
    $(id).addEventListener("change", renderComparison);
  $("reset").onclick = () => {
    $("cohort").value = "full";
    for (const id of [
      "dataset",
      "category",
      "tier",
      "model-search",
      "question-search",
    ])
      $(id).value = "";
    $("question-filter").value = "all";
    $("effort-pair").value = "medium|max";
    selectedQuestion = "";
    sortKey = "accuracy";
    ascending = false;
    render();
  };
  $("print").onclick = () => doc.defaultView.print();
  renderProvenance();
  render();
  return { render };
}
if (typeof module !== "undefined" && module.exports)
  module.exports = {
    matchesScope,
    compareOutcomes,
    sortedRows,
    difficultyBins,
    rankSpan,
    separationExtreme,
    initReport,
  };
if (typeof document !== "undefined" && document.getElementById("data")) {
  const start = () =>
    initReport(
      document,
      JSON.parse(document.getElementById("data").textContent),
    );
  // Font metrics must be ready before wrapping chart captions and laying out labels.
  if (document.fonts)
    document.fonts.load("500 16px Inter").then(start, (err) => {
      console.error(
        "Report typography could not load; using the system fallback.",
        err,
      );
      start();
    });
  else start();
}
