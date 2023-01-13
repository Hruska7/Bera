// Copyright (C) 2023, Berachain Foundation. All rights reserved.
// See the file LICENSE for licensing terms.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package mock

import (
	"github.com/berachain/stargazer/core/state/store/journal"
)

// `MockCacheEntry` is a basic CacheEntry which increases num by 1 on `Revert()`.
type CacheEntry struct {
	num int
}

// `NewCacheEntry` creates a new `MockCacheEntry`.
func NewCacheEntry() *CacheEntry {
	return &CacheEntry{}
}

// `Revert` implements `CacheEntry`.
func (m *CacheEntry) Revert() {
	m.num++
}

// `Clone` implements `CacheEntry`.
func (m *CacheEntry) Clone() journal.CacheEntry {
	return &CacheEntry{num: m.num}
}

// `RevertCallCount` returns the number of times `Revert` has been called.
func (m *CacheEntry) RevertCallCount() int {
	return m.num
}
