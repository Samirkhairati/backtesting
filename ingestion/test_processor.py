from datetime import date

from processor import missing_segments

D = date


def test_no_coverage_downloads_everything():
    assert missing_segments(D(2019, 1, 1), D(2023, 12, 31), None) == [(D(2019, 1, 1), D(2023, 12, 31))]


def test_fully_covered_downloads_nothing():
    assert missing_segments(D(2020, 1, 1), D(2022, 1, 1), (D(2019, 1, 1), D(2023, 12, 31))) == []


def test_extension_to_the_right_only_fetches_the_delta():
    assert missing_segments(D(2019, 1, 1), D(2026, 10, 15), (D(2019, 1, 1), D(2023, 12, 31))) == [(D(2023, 12, 31), D(2026, 10, 15))]


def test_extension_to_the_left_only_fetches_the_delta():
    assert missing_segments(D(2012, 1, 1), D(2023, 1, 1), (D(2019, 1, 1), D(2023, 12, 31))) == [(D(2012, 1, 1), D(2019, 1, 1))]


def test_both_sides_fetch_two_deltas():
    assert missing_segments(D(2012, 1, 1), D(2026, 1, 1), (D(2019, 1, 1), D(2023, 12, 31))) == [
        (D(2012, 1, 1), D(2019, 1, 1)),
        (D(2023, 12, 31), D(2026, 1, 1)),
    ]


def test_request_after_the_envelope_fills_the_gap_so_coverage_stays_truthful():
    # Asking for 2025 when only up to 2023 is stored must also fetch 2024.
    assert missing_segments(D(2025, 1, 1), D(2025, 6, 1), (D(2019, 1, 1), D(2023, 12, 31))) == [(D(2023, 12, 31), D(2025, 6, 1))]
