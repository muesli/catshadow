// constants.go - mixnet client constants
// Copyright (C) 2018  David Stainton.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package constants

const (
	// MessageIDLength is the length of a message ID in bytes.
	MessageIDLength = 16
)

const (
	// SurbTypeACK is used to denote an ACK in response to a forward message.
	SurbTypeACK = 0

	// SurbTypeKaetzchen is used to denote a mixnet service query response.
	SurbTypeKaetzchen = 1

	// SurbTypeInternal is used to reserve an internal SURB reply type.
	SurbTypeInternal = 2
)
