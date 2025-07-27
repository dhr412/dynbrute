# DynBrute

`DynBrute` is a CLI tool built in Go for brute-forcing web login credentials. This tool can use either provided wordlists for usernames and passwords or generate random credentials, and it attempts to log in, checking for successful authentication based on page redirection. It leverages parallelism to efficiently test large numbers of credential combinations.

> Ensure that Google Chrome or Chromium are installed and accessible in your system's PATH for `DynBrute` to work.

---

## Features

* **Concurrent Login Attempts** – Utilizes goroutines to perform multiple login attempts simultaneously, with a maximum of 256 parallel requests.
* **Credential Sources** – Supports optional username and password wordlists (.txt files) or falls back to random credential generation (1-32 characters).
* **Browser Automation** – Uses the `chromedp` library to simulate real user interactions with the login form.
* **Success Detection** – Determines successful logins by checking if the page redirects to a different URL after form submission.
* **Flexible Configuration** – Allows specification of the target URL, wordlist files, and number of attempts via command-line flags.

---

## Installation

### Download from Releases

1. Visit the [Releases](https://github.com/dhr412/dynbrute/releases) page.
2. Download the binary for your platform.
3. Make it executable (on Unix-like systems):

   ```bash
   chmod +x dynbrute
   ```

4. Run it with

    ```bash
    ./dynbrute -help
    ```

### Compiling from Source

Ensure you have [Go](https://go.dev/dl) installed:

1. Clone the DynBrute repository:

   ```bash
   git clone https://github.com/dhr412/dynbrute.git
   ```

2. Navigate to the project directory:

   ```bash
   cd dynbrute
   ```

3. Build the executable:

   ```bash
   go build -o dynbrute
   ```

4. Run it with:

   ```bash
   ./dynbrute --help
   ```

---

## Usage

```bash
dynbrute -url <target_url> [-users <usernames.txt>] [-passwords <passwords.txt>] [-attempts <num_attempts>]
```

### Required Flags

* `-url`: Target website login URL (e.g., `http://example.com/login`)

### Optional Flags

* `-users`: Path to usernames wordlist (.txt file)
* `-passwords`: Path to passwords wordlist (.txt file)
* `-attempts`: Number of login attempts (default: 128)

### Examples

* With wordlists:

  ```bash
  dynbrute -url http://example.com/login -users usernames.txt -passwords passwords.txt -attempts 100
  ```

* Without wordlists (falls back to random generation):

  ```bash
  dynbrute -url http://example.com/login -attempts 100
  ```

---

## How It Works

1. **Input Parsing**:
   * The tool parses command-line flags to retrieve the target URL, optional wordlist files, and number of login attempts.

2. **URL Normalization**:
   * If the URL doesn’t start with `http://` or `https://`, it prepends `http://` to ensure proper navigation.

3. **Credential Sources**:
   * If both username and password wordlists are provided and valid (.txt files with content), it randomly selects credentials from them.
   * If either wordlist is missing or invalid, it falls back to generating random usernames and passwords (1-32 characters).

4. **Task Generation**:
   * Creates tasks with either wordlist-based or randomly generated credentials.

5. **Concurrent Execution**:
   * Uses goroutines and a semaphore to limit concurrent attempts to 256, preventing server or local machine overload.

6. **Browser Automation**:
   * For each task, the `chromedp` library:
     * Navigates to the login page.
     * Waits for the username (`input[type="text"]`) and password (`input[type="password"]`) fields to be visible.
     * Fills in the credentials.
     * Clicks the submit button (`input[type="submit"]`).
     * Waits 8 seconds to check for redirection.
     * Retrieves the current URL.

7. **Success Check**:
   * If the current URL differs from the initial login URL, it logs the attempt as successful with the credentials and redirected URL.
   * Otherwise, it logs the attempt as a failure.

---

## License

This project is licensed under the MIT license.

