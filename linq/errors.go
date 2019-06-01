/*
adammil.net/linq is a library that implements .NET-like LINQ queries for Go.

http://www.adammil.net/
Copyright (C) 2019 Adam Milazzo

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package linq

// Determines whether the given error indicates that a sequence was empty or no items matched a predicate.
func IsEmptyError(e error) bool {
	_, ok := e.(emptyError)
	return ok
}

// Determines whether the given error indicates that a sequence had too many or too many items matched a predicate.
func IsTooManyItemsError(e error) bool {
	_, ok := e.(tooManyItemsError)
	return ok
}

type emptyError struct{}

func (emptyError) Error() string { return "the sequence was empty (or no items matched)" }

type tooManyItemsError struct{}

func (tooManyItemsError) Error() string {
	return "the sequence contained too many items (or too many items matched)"
}
