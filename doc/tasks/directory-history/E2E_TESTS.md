# E2E Test Scenarios: Directory History Navigation

## Test Environment

**Build command:**
```bash
make test-e2e-build
```

**Run command:**
```bash
make test-e2e  # Run all automated tests
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "{commands}"  # Interactive
```

## Scenario 1: 基本的な履歴ナビゲーション（前に戻る）

**Purpose:** ディレクトリを移動後、`[`キーで履歴を遡って戻れることを確認

**Preconditions:**
- テスト環境にディレクトリ構造が存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter WAIT [ WAIT C-c"
```

**Command Breakdown:**
1. `j` - カーソルを下に移動
2. `j` - さらに下に移動
3. `Enter` - ディレクトリに入る
4. `WAIT` - ディレクトリ読み込みを待つ
5. `[` - 履歴を遡って前のディレクトリに戻る
6. `WAIT` - ディレクトリ読み込みを待つ
7. `C-c` - アプリケーション終了

**Expected Results:**
- ディレクトリに入った後、`[`キーで元のディレクトリに戻る
- 履歴ナビゲーション後、正しいディレクトリが表示される
- エラーメッセージが表示されない

**Verification Method:**
- 画面キャプチャで最終状態を確認
- ディレクトリパスが元の位置に戻っていることを確認
- クリーンな終了（クラッシュなし）

**Success Criteria:**
- ✅ 履歴を遡って正しいディレクトリに戻れた
- ✅ エラーなく動作した
- ✅ 画面状態が期待通り

---

## Scenario 2: 履歴の前進（Forward）

**Purpose:** 履歴を遡った後、`]`キーで再び進めることを確認

**Preconditions:**
- テスト環境にディレクトリ構造が存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter WAIT [ WAIT ] WAIT C-c"
```

**Command Breakdown:**
1. `j j Enter` - ディレクトリに移動
2. `WAIT` - 読み込み待機
3. `[` - 履歴を遡る
4. `WAIT` - 読み込み待機
5. `]` - 履歴を進む
6. `WAIT` - 読み込み待機
7. `C-c` - 終了

**Expected Results:**
- `[`で戻った後、`]`で再びサブディレクトリに進める
- 履歴の前進/後退がスムーズに動作する
- エラーメッセージなし

**Verification Method:**
- 最終的にサブディレクトリに戻っていることを確認
- 画面キャプチャでディレクトリパスを検証

**Success Criteria:**
- ✅ 履歴の前進が正常に動作
- ✅ ディレクトリが正しく切り替わった
- ✅ エラーなし

---

## Scenario 3: 複数階層のナビゲーション

**Purpose:** 複数のディレクトリを移動後、履歴で複数回戻れることを確認

**Preconditions:**
- テスト環境に3階層以上のディレクトリ構造が存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter WAIT j j Enter WAIT [ WAIT [ WAIT C-c"
```

**Command Breakdown:**
1. `j j Enter` - 第1階層のディレクトリに入る
2. `WAIT` - 読み込み待機
3. `j j Enter` - 第2階層のディレクトリに入る
4. `WAIT` - 読み込み待機
5. `[` - 1回戻る（第1階層へ）
6. `WAIT` - 読み込み待機
7. `[` - もう1回戻る（最初の場所へ）
8. `WAIT` - 読み込み待機
9. `C-c` - 終了

**Expected Results:**
- 2回の`[`操作で元の位置まで戻れる
- 各ステップで正しいディレクトリが表示される
- 履歴が正しく管理されている

**Verification Method:**
- 最終的に最初のディレクトリに戻っていることを確認
- 画面キャプチャでパス履歴を検証

**Success Criteria:**
- ✅ 複数階層の履歴ナビゲーションが動作
- ✅ 各階層で正しいディレクトリが表示された
- ✅ エラーなし

---

## Scenario 4: 履歴がない状態での操作

**Purpose:** 起動直後など履歴がない状態で`[`や`]`を押しても問題ないことを確認

**Preconditions:**
- アプリケーション起動直後

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "[ ] C-c"
```

**Command Breakdown:**
1. `[` - 履歴を遡る（履歴なし）
2. `]` - 履歴を進む（履歴なし）
3. `C-c` - 終了

**Expected Results:**
- エラーメッセージが表示されない
- アプリケーションがクラッシュしない
- 何も起こらず、同じ場所に留まる

**Verification Method:**
- エラーメッセージがないことを確認
- クリーンな終了

**Success Criteria:**
- ✅ 履歴がなくてもエラーなし
- ✅ アプリケーションが安定動作
- ✅ 画面状態が変わらない

---

## Scenario 5: `-`キーとの独立動作

**Purpose:** 履歴ナビゲーションと`-`キー（previousPath）が独立して動作することを確認

**Preconditions:**
- テスト環境にディレクトリ構造が存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter WAIT j j Enter WAIT - WAIT - WAIT [ WAIT C-c"
```

**Command Breakdown:**
1. `j j Enter` - ディレクトリAに入る
2. `WAIT` - 読み込み待機
3. `j j Enter` - ディレクトリBに入る
4. `WAIT` - 読み込み待機
5. `-` - previousPathでディレクトリAに戻る
6. `WAIT` - 読み込み待機
7. `-` - 再びディレクトリBに戻る（トグル動作）
8. `WAIT` - 読み込み待機
9. `[` - 履歴を遡る（ディレクトリAに戻る）
10. `WAIT` - 読み込み待機
11. `C-c` - 終了

**Expected Results:**
- `-`キーでトグル動作が正常に機能
- `[`キーで履歴を遡ることもできる
- 両機能が独立して動作する

**Verification Method:**
- `-`キーと`[`キーの動作を確認
- 最終的に正しいディレクトリにいることを確認

**Success Criteria:**
- ✅ `-`キーのトグル動作が正常
- ✅ 履歴ナビゲーションも正常
- ✅ 両機能が干渉しない

---

## Scenario 6: 履歴の途中で新しいディレクトリに移動

**Purpose:** 履歴を遡った後、新しいディレクトリに移動すると前方履歴がクリアされることを確認

**Preconditions:**
- テスト環境に複数のディレクトリが存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter WAIT j j Enter WAIT [ WAIT k Enter WAIT ] C-c"
```

**Command Breakdown:**
1. `j j Enter` - ディレクトリAに入る
2. `WAIT` - 読み込み待機
3. `j j Enter` - ディレクトリBに入る
4. `WAIT` - 読み込み待機
5. `[` - ディレクトリAに戻る
6. `WAIT` - 読み込み待機
7. `k Enter` - 別のディレクトリCに入る
8. `WAIT` - 読み込み待機
9. `]` - 前方履歴を試す（すでにクリアされているはず）
10. `C-c` - 終了

**Expected Results:**
- 新しいディレクトリに移動後、`]`キーは効果なし（前方履歴がクリアされた）
- ディレクトリCに留まる
- エラーメッセージなし

**Verification Method:**
- `]`キー押下後もディレクトリCにいることを確認
- 前方履歴が正しくクリアされたことを確認

**Success Criteria:**
- ✅ 前方履歴が正しくクリアされた
- ✅ 新しいディレクトリへの移動が正常
- ✅ エラーなし

---

## Scenario 7: 親ディレクトリ移動も履歴に記録

**Purpose:** `h`キーや`..`での親ディレクトリ移動も履歴に記録されることを確認

**Preconditions:**
- テスト環境にディレクトリ構造が存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter WAIT h WAIT [ WAIT C-c"
```

**Command Breakdown:**
1. `j j Enter` - サブディレクトリに入る
2. `WAIT` - 読み込み待機
3. `h` - 親ディレクトリに戻る
4. `WAIT` - 読み込み待機
5. `[` - 履歴を遡る（サブディレクトリに戻る）
6. `WAIT` - 読み込み待機
7. `C-c` - 終了

**Expected Results:**
- 親ディレクトリ移動も履歴に記録される
- `[`でサブディレクトリに戻れる
- 正常に動作する

**Verification Method:**
- 最終的にサブディレクトリにいることを確認
- 親ディレクトリ移動が履歴に記録されたことを確認

**Success Criteria:**
- ✅ 親ディレクトリ移動が履歴に記録された
- ✅ 履歴ナビゲーションで戻れた
- ✅ エラーなし

---

## Scenario 8: ホームディレクトリ移動も履歴に記録

**Purpose:** `~`キーでのホームディレクトリ移動も履歴に記録されることを確認

**Preconditions:**
- テスト環境が存在する

**Test Commands:**
```bash
docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "~ WAIT [ WAIT C-c"
```

**Command Breakdown:**
1. `~` - ホームディレクトリに移動
2. `WAIT` - 読み込み待機
3. `[` - 履歴を遡る（元のディレクトリに戻る）
4. `WAIT` - 読み込み待機
5. `C-c` - 終了

**Expected Results:**
- ホームディレクトリ移動が履歴に記録される
- `[`で元のディレクトリに戻れる
- エラーなし

**Verification Method:**
- 最終的に元のディレクトリに戻っていることを確認
- ホームディレクトリ移動が履歴に記録されたことを確認

**Success Criteria:**
- ✅ ホームディレクトリ移動が履歴に記録された
- ✅ 履歴ナビゲーションで戻れた
- ✅ エラーなし

---

## Test Execution Checklist

- [ ] Build E2E environment: `make test-e2e-build`
- [ ] Run automated tests: `make test-e2e`
- [ ] Execute Scenario 1: 基本的な履歴ナビゲーション
- [ ] Execute Scenario 2: 履歴の前進
- [ ] Execute Scenario 3: 複数階層のナビゲーション
- [ ] Execute Scenario 4: 履歴がない状態での操作
- [ ] Execute Scenario 5: `-`キーとの独立動作
- [ ] Execute Scenario 6: 履歴の途中で新しいディレクトリに移動
- [ ] Execute Scenario 7: 親ディレクトリ移動も履歴に記録
- [ ] Execute Scenario 8: ホームディレクトリ移動も履歴に記録
- [ ] Analyze all results
- [ ] Document failures (if any)

## Additional Manual Tests

以下は自動化が難しいため手動でテストすることを推奨：

1. **Alt+Left / Alt+Right キーの動作**
   - 端末によって認識が異なる可能性があるため手動確認

2. **100エントリの履歴上限**
   - 100以上のディレクトリ移動を行い、古い履歴が削除されることを確認

3. **パフォーマンステスト**
   - 高速な履歴ナビゲーション操作でも問題ないことを確認

4. **左右ペイン独立性**
   - 左ペインと右ペインがそれぞれ独立した履歴を持つことを確認
