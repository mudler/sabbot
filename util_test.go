package main

import (
    "strings"
    "testing"
)

func TestSearchCmd(t *testing.T) {
    topicalResult := SearchCmd("New York")
    if !strings.Contains(topicalResult, "First Topical Result:") {
        t.Errorf("SearchCmd(%v), got %v, want 'First Topical Result:'\n", "New York", topicalResult)
    }
    redirectResult := SearchCmd("!archwiki i3")
    if !strings.Contains(redirectResult, "Redirect result:") {
        t.Errorf("SearchCmd(%v), got %v, want 'Redirect result:'\n", "!archwiki i3", redirectResult)
    }
    noResult := SearchCmd("my face")
    if !strings.Contains(noResult, "returned no results") {
        t.Errorf("SearchCmd(%v), got %v, want 'Query: my face returned no results.'\n", "your face", noResult)
    }
}

func TestUrlTitle(t *testing.T) {
    result := UrlTitle("www.exploit-db.com")
     if !strings.Contains(result, "Exploits Database by Offensive Security") {
        t.Errorf("UrlTitle(%v), got %v, want '[ Exploits Database by Offensive Security ]( http://www.exploit-db.com/ )'\n", "www.exploit-db.com", result)
    }
}